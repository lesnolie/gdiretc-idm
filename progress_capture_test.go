package main

import (
 "strings"
 "testing"
)
func TestCaptureLibaria2Speed(t *testing.T){
 var got float64
 out:=&progressCapture{onSpeed:func(v float64){got=v}}
 for _,chunk:=range []string{"libaria2 ready\n\rDownlo","ading: 128 KiB/s\r","Downloading: 256 KiB/s\n"}{
  if _,err:=out.Write([]byte(chunk));err!=nil{t.Fatal(err)}
 }
 if got!=256*1024{t.Fatalf("download speed=%v, want %d",got,256*1024)}
 if !strings.Contains(out.String(),"Downloading: 256 KiB/s"){t.Fatal("diagnostic output was lost")}
}
