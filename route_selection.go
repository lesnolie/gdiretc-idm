package main

import (
 "context"
 "errors"
 "fmt"
 "net"
 "os/exec"
 "sort"
 "strconv"
 "strings"
 "time"
)

// RouteOption contains a gdirect-scanned Google endpoint and its latency.
// Latency is a connection probe, not a guarantee of download throughput.
type RouteOption struct {
 IP string
 SNI string
 LatencyMS int
}

func (s *Service) ListRoutes() ([]RouteOption,error){
 if s.helperErr!=nil{return nil,s.helperErr}
 ctx,cancel:=context.WithTimeout(context.Background(),15*time.Second)
 defer cancel()
 cmd:=exec.CommandContext(ctx,s.helper,"scan","--list")
 hideHelperWindow(cmd)
 out,err:=cmd.Output()
 if err!=nil{return nil,fmt.Errorf("reading cached Google IPs: %w",err)}
 endpoints:=parseSavedEndpoints(string(out))
 routes:=make([]RouteOption,0,len(endpoints))
 for _,ep:=range endpoints{
  if ep.latencyMS<0{continue}
  routes=append(routes,RouteOption{IP:ep.ip,SNI:ep.sni,LatencyMS:ep.latencyMS})
 }
 sort.SliceStable(routes,func(i,j int)bool{return routes[i].LatencyMS<routes[j].LatencyMS})
 if len(routes)>16{routes=routes[:16]}
 return routes,nil
}

func (s *Service) RefreshRoutes() ([]RouteOption,error){
 if s.helperErr!=nil{return nil,s.helperErr}
 ctx,cancel:=context.WithTimeout(context.Background(),120*time.Second)
 defer cancel()
 cmd:=exec.CommandContext(ctx,s.helper,"scan")
 hideHelperWindow(cmd)
 if out,err:=cmd.CombinedOutput();err!=nil{
  msg:=strings.TrimSpace(string(out))
  if len(msg)>280{msg=msg[len(msg)-280:]}
  return nil,fmt.Errorf("Google IP scan failed: %w — %s",err,msg)
 }
 return s.ListRoutes()
}

func checkRoute(ip,sni string)error{
 if ip==""{
  if sni!=""{return errors.New("automatic route cannot specify a TLS identity")}
  return nil
 }
 if net.ParseIP(ip)==nil{return errors.New("invalid Google IP")}
 if sni!=""&&sni!="g.cn"{return errors.New("unsupported TLS identity")}
 return nil
}
func routeArgs(ip,sni string)[]string{
 if ip==""{return nil}
 args:=[]string{"--ip",ip}
 if sni!=""{args=append(args,"--sni",sni)}
 return args
}
func (s *Service) ResolveRoute(ip string)(RouteOption,error){
 if ip==""{return RouteOption{},nil}
 if net.ParseIP(ip)==nil{return RouteOption{},errors.New("invalid IP address")}
 routes,err:=s.ListRoutes()
 if err!=nil{return RouteOption{},err}
 for _,r:=range routes{
  if r.IP==ip{return r,nil}
 }
 return RouteOption{},errors.New("IP not found among cached verified Google routes; rescan first")
}
func routeLabel(route RouteOption)string{
 display:=route.IP+" · "+strconv.Itoa(route.LatencyMS)+" ms"
 if route.SNI!=""{display+=" · SNI "+route.SNI}
 return display
}
