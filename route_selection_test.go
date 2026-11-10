package main

import (
 "net/http"
 "reflect"
 "testing"
)

func TestManualRouteArgs(t *testing.T){
 cases:=[]struct{ip,sni string;args []string}{
  {"","",nil},
  {"142.250.2.1","",[]string{"--ip","142.250.2.1"}},
  {"142.250.2.1","g.cn",[]string{"--ip","142.250.2.1","--sni","g.cn"}},
 }
 for _,tc:=range cases{
  if err:=checkRoute(tc.ip,tc.sni);err!=nil{t.Fatal(err)}
  if got:=routeArgs(tc.ip,tc.sni);!reflect.DeepEqual(got,tc.args){t.Errorf("routeArgs: got %v want %v",got,tc.args)}
 }
 if checkRoute("127.0.0.1","evil.com")==nil{t.Fatal("invalid SNI allowed")}
 if checkRoute("not-an-ip","")==nil{t.Fatal("invalid IP allowed")}
 if checkRoute("","g.cn")==nil{t.Fatal("automatic SNI override allowed")}
}
func TestProgressRatio(t *testing.T){
 cases:=[]struct{task Task; want float64}{
  {Task{State:"running",Bytes:25,Total:100},0.25},
  {Task{State:"running",Bytes:50,Total:100},0.5},
  {Task{State:"running",Bytes:100,Total:100},0.999},
  {Task{State:"complete",Bytes:100,Total:100},1},
  {Task{State:"running",Bytes:30},-1},
 }
 for _,tc:=range cases{
  got:=progress(tc.task)
  if got!=tc.want{t.Errorf("progress(%+v)=%v want %v",tc.task,got,tc.want)}
 }
}
func TestTotalSizeFromRange(t *testing.T){
 h:=http.Header{}
 h.Set("Content-Range","bytes 0-0/1048576")
 if got:=sizeFromHeaders(h,http.StatusPartialContent);got!=1048576{
  t.Fatalf("content-range total=%d",got)
 }
}
