package main

import (
 "regexp"
 "strconv"
 "strings"
 "sync"
)

var gdirectSpeedLine=regexp.MustCompile(`Downloading:\s*([0-9]+)\s+KiB/s`)

// gdirect emits real libaria2 download speed to stderr every ~100ms.
// Capture it separately from the ring-buffered diagnostic messages.
type progressCapture struct{
 mu sync.Mutex
 log limitedBuffer
 tail string
 onSpeed func(float64)
}
func (p *progressCapture) Write(chunk []byte)(int,error){
 p.mu.Lock()
 _,_ = p.log.Write(chunk)
 text:=p.tail+string(chunk)
 matches:=gdirectSpeedLine.FindAllStringSubmatch(text,-1)
 if len(text)>160{p.tail=text[len(text)-160:]}else{p.tail=text}
 p.mu.Unlock()
 if len(matches)>0&&p.onSpeed!=nil{
  last:=matches[len(matches)-1][1]
  if value,err:=strconv.ParseInt(strings.TrimSpace(last),10,64);err==nil{
   p.onSpeed(float64(value)*1024)
  }
 }
 return len(chunk),nil
}
func (p *progressCapture) String()string{
 p.mu.Lock()
 defer p.mu.Unlock()
 return p.log.String()
}
