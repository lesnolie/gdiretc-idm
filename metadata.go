package main

import (
 "context"
 "crypto/tls"
 "errors"
 "fmt"
 "html"
 "io"
 "mime"
 "net"
 "net/http"
 "net/url"
 "os/exec"
 "path/filepath"
 "regexp"
 "strings"
 "time"

 "github.com/lesnolie/gdiretc-idm/internal/driveurl"
)

type savedEndpoint struct {ip string; sni string}
var (
 metaTag = regexp.MustCompile(`(?is)<meta\b[^>]*>`)
 attribute = regexp.MustCompile(`(?is)([a-zA-Z0-9_:-]+)\s*=\s*(?:"([^"]*)"|'([^']*)')`)
)
func parseSavedEndpoints(raw string) []savedEndpoint {
 var entries []savedEndpoint
 for _,line:=range strings.Split(raw,"\n") {
  parts:=strings.Split(strings.TrimSpace(line),"\t")
  if len(parts)<3 || parts[1]!="OK" {continue}
  ip:=net.ParseIP(parts[0]);if ip==nil{continue}
  sni:=""
  if strings.Contains(parts[2],"(SNI g.cn)") {sni="g.cn"}
  entries=append(entries,savedEndpoint{ip:ip.String(),sni:sni})
 }
 return entries
}
func allowedMetadataHost(host string) bool {
 host=strings.ToLower(host)
 return host=="drive.google.com" || host=="drive.usercontent.google.com" || strings.HasSuffix(host,".googleusercontent.com")
}
func publicMetadataClient(endpoint savedEndpoint) *http.Client {
 transport:=&http.Transport{
  Proxy:nil,
  DialContext:func(ctx context.Context, network, address string)(net.Conn,error){
   host,port,err:=net.SplitHostPort(address);if err!=nil{return nil,err}
   if port!="443"||!allowedMetadataHost(host){return nil,errors.New("metadata redirect is not a Google download host")}
   return (&net.Dialer{Timeout:6*time.Second}).DialContext(ctx,"tcp",net.JoinHostPort(endpoint.ip,"443"))
  },
  TLSHandshakeTimeout:6*time.Second,
  DisableKeepAlives:true,
 }
 if endpoint.sni!=""{
  transport.TLSClientConfig=&tls.Config{MinVersion:tls.VersionTLS12,ServerName:endpoint.sni}
 }
 return &http.Client{
  Transport:transport,Timeout:15*time.Second,
  CheckRedirect:func(req *http.Request,via []*http.Request)error{
   if len(via)>5||req.URL.Scheme!="https"||!allowedMetadataHost(req.URL.Hostname()){return errors.New("untrusted Google Drive redirect")}
   return nil
  },
 }
}
func safeName(raw string) string {
 raw=html.UnescapeString(strings.TrimSpace(raw))
 raw=strings.ReplaceAll(raw,"\\","/")
 raw=filepath.Base(raw)
 raw=strings.Map(func(r rune)rune{if r<32||r==127||strings.ContainsRune(`<>:"/\\|?*`,r){return -1};return r},raw)
 raw=strings.Trim(raw," .")
 if raw=="."||raw==".."||raw==""{return ""}
 if len([]rune(raw))>200{return ""}
 return raw
}
func nameFromDisposition(value string) string{
 _,params,err:=mime.ParseMediaType(value);if err!=nil{return ""}
 return safeName(params["filename"])
}
func nameFromHTML(raw string) string{
 for _,tag:=range metaTag.FindAllString(raw,-1){
  attrs:=map[string]string{}
  for _,a:=range attribute.FindAllStringSubmatch(tag,-1){
   value:=a[2];if value==""{value=a[3]}
   attrs[strings.ToLower(a[1])]=value
  }
  p:=attrs["property"]
  if p=="og:title"||attrs["itemprop"]=="name"{
   name:=safeName(attrs["content"])
   if name!=""&&name!="Google Drive"&&name!="Sign in"{return name}
  }
 }
 return ""
}
func fetchPublicName(ctx context.Context, ep savedEndpoint, id,key string)(string,error){
 client:=publicMetadataClient(ep)
 q:=url.Values{"id":{id},"export":{"download"},"confirm":{"t"}}
 if key!=""{q.Set("resourcekey",key)}
 urls:=[]string{"https://drive.usercontent.google.com/download?"+q.Encode()}
 q2:=url.Values{}
 if key!=""{q2.Set("resourcekey",key)}
 viewURL:="https://drive.google.com/file/d/"+url.PathEscape(id)+"/view"
 if len(q2)>0{viewURL+="?"+q2.Encode()}
 urls=append(urls,viewURL)
 for index,target:=range urls{
  req,err:=http.NewRequestWithContext(ctx,http.MethodGet,target,nil);if err!=nil{return "",err}
  if index==0{req.Header.Set("Range","bytes=0-0")}
  resp,err:=client.Do(req);if err!=nil{continue}
  if resp.StatusCode>=200&&resp.StatusCode<300{
   if name:=nameFromDisposition(resp.Header.Get("Content-Disposition"));name!=""{resp.Body.Close();return name,nil}
   if strings.Contains(strings.ToLower(resp.Header.Get("Content-Type")),"text/html"){
    body,_:=io.ReadAll(io.LimitReader(resp.Body,2<<20))
    if name:=nameFromHTML(string(body));name!=""{resp.Body.Close();return name,nil}
   }
  }
  resp.Body.Close()
 }
 return "",errors.New("public file did not expose its original filename")
}
func (s *Service) SuggestName(raw string)(string,error){
 ref,err:=driveurl.Parse(strings.TrimSpace(raw));if err!=nil{return "",err}
 ctx,cancel:=context.WithTimeout(context.Background(),100*time.Second);defer cancel()
 list:=func()[]savedEndpoint{
  command:=exec.CommandContext(ctx,s.helper,"scan","--list")
  stdout,err:=command.Output()
  if err!=nil{return nil}
  return parseSavedEndpoints(string(stdout))
 }
 endpoints:=list()
 if len(endpoints)==0{
  if out,err:=exec.CommandContext(ctx,s.helper,"scan").CombinedOutput();err!=nil{
   return "",fmt.Errorf("IP scan failed: %s (%v)",string(out),err)
  }
  endpoints=list()
 }
 if len(endpoints)==0{return "",errors.New("no verified direct Google IP available")}
 for i,ep:=range endpoints{
  if i>=3{break}
  name,err:=fetchPublicName(ctx,ep,ref.ID,ref.ResourceKey)
  if err==nil&&name!=""{return name,nil}
 }
 return "",errors.New("cannot read original filename from this public link; please choose a name manually")
}
