package main

import (
 "image"
 "image/color"
 "image/png"
 "math"
 "os"
 "path/filepath"
)

func render(size int)*image.NRGBA{
 img:=image.NewNRGBA(image.Rect(0,0,size,size))
 blue:=color.NRGBA{R:32,G:103,B:220,A:255}
 white:=color.NRGBA{R:255,G:255,B:255,A:255}
 for y:=0;y<size;y++{for x:=0;x<size;x++{
  nx:=(float64(x)+0.5)/float64(size)
  ny:=(float64(y)+0.5)/float64(size)
  // Rounded-square background, transparent outside.
  dx:=math.Max(math.Abs(nx-0.5)-0.31,0)
  dy:=math.Max(math.Abs(ny-0.5)-0.31,0)
  if dx*dx+dy*dy>0.19*0.19{continue}
  img.SetNRGBA(x,y,blue)
  // A deliberately geometric G: circular stroke, right opening, and crossbar.
  xx,yy:=nx-0.50,ny-0.50
  r:=math.Hypot(xx,yy)
  ring:=r>=0.235&&r<=0.335
  gap:=xx>0.12&&yy>-0.10&&yy<0.08
  bar:=nx>=0.51&&nx<=0.79&&ny>=0.485&&ny<=0.585
  right:=nx>=0.675&&nx<=0.79&&ny>=0.485&&ny<=0.70
  if (ring&&!gap)||bar||right{img.SetNRGBA(x,y,white)}
 }}
 return img
}
func save(path string,img image.Image)error{
 if err:=os.MkdirAll(filepath.Dir(path),0755);err!=nil{return err}
 f,err:=os.Create(path);if err!=nil{return err}
 defer f.Close()
 return png.Encode(f,img)
}
func main(){
 if err:=save("resources/icon.png",render(1024));err!=nil{panic(err)}
 if err:=save("resources/tray.png",render(64));err!=nil{panic(err)}
}
