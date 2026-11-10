package driveurl

import (
 "errors"
 "net/url"
 "regexp"
 "strings"
)
var (
 ErrInvalid=errors.New("not a supported Google Drive file link")
 ErrFolder=errors.New("Google Drive folder links are not supported")
 idPattern=regexp.MustCompile(`^[A-Za-z0-9_-]{10,}$`)
)
type File struct { ID string; ResourceKey string }
func Parse(raw string) (File,error) {
 u,err:=url.Parse(strings.TrimSpace(raw))
 if err!=nil||u.Scheme!="https"||u.User!=nil||u.Port()!="" {return File{},ErrInvalid}
 host:=strings.ToLower(u.Hostname())
 if host!="drive.google.com"&&host!="drive.usercontent.google.com" {return File{},ErrInvalid}
 path,err:=url.PathUnescape(strings.Trim(u.EscapedPath(),"/"))
 if err!=nil {return File{},ErrInvalid}
 parts:=strings.Split(path,"/")
 if len(parts)>0&&(parts[0]=="folders"||(len(parts)>1&&parts[0]=="drive"&&parts[1]=="folders")){return File{},ErrFolder}
 var id string
 switch {
 case host=="drive.google.com"&&len(parts)>=3&&parts[0]=="file"&&parts[1]=="d":
  if len(parts)>3&&(len(parts)!=4||(parts[3]!="view"&&parts[3]!="edit")) {return File{},ErrInvalid}
  id=parts[2]
 case host=="drive.google.com"&&(path=="open"||path=="uc"):
  id=u.Query().Get("id")
 case host=="drive.usercontent.google.com"&&path=="download":
  id=u.Query().Get("id")
 default: return File{},ErrInvalid
 }
 if !idPattern.MatchString(id){return File{},ErrInvalid}
 return File{ID:id,ResourceKey:u.Query().Get("resourcekey")},nil
}
