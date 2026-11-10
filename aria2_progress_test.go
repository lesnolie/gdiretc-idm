package main

import (
 "encoding/binary"
 "os"
 "path/filepath"
 "testing"
)

func writeAria2Fixture(path string,total uint64,piece uint32,field []byte)error{
 data:=make([]byte,34+len(field)+4)
 binary.BigEndian.PutUint16(data[0:2],1)
 binary.BigEndian.PutUint32(data[2:6],0)
 binary.BigEndian.PutUint32(data[6:10],0)
 binary.BigEndian.PutUint32(data[10:14],piece)
 binary.BigEndian.PutUint64(data[14:22],total)
 binary.BigEndian.PutUint64(data[22:30],0)
 binary.BigEndian.PutUint32(data[30:34],uint32(len(field)))
 copy(data[34:],field)
 return os.WriteFile(path+".aria2",data,0600)
}

func TestCompletedPiecesNotSparseFileSize(t *testing.T){
 dir:=t.TempDir()
 file:=filepath.Join(dir,"video.mp4")
 // The output can have its full logical length even when aria2 only
 // downloaded the first two 1-MiB pieces.
 if err:=os.WriteFile(file,[]byte{1},0600);err!=nil{t.Fatal(err)}
 if err:=os.Truncate(file,10<<20);err!=nil{t.Fatal(err)}
 if err:=writeAria2Fixture(file,10<<20,1<<20,[]byte{0b11000000,0});err!=nil{t.Fatal(err)}
 done,total,err:=readAria2Progress(file)
 if err!=nil{t.Fatal(err)}
 if done!=2<<20||total!=10<<20{t.Fatalf("got %d/%d, want 2MiB/10MiB",done,total)}
}

func TestLastPartialPiece(t *testing.T){
 path:=filepath.Join(t.TempDir(),"sample")
 if err:=writeAria2Fixture(path,(2<<20)+(512<<10),1<<20,[]byte{0b10100000});err!=nil{t.Fatal(err)}
 got,total,err:=readAria2Progress(path)
 if err!=nil{t.Fatal(err)}
 if got!=(1<<20)+(512<<10)||total!=(2<<20)+(512<<10){t.Fatalf("%d/%d",got,total)}
}
func TestRejectShortControlFile(t *testing.T){
 path:=filepath.Join(t.TempDir(),"bad")
 if err:=os.WriteFile(path+".aria2",[]byte{0,1,1},0600);err!=nil{t.Fatal(err)}
 if _,_,err:=readAria2Progress(path);err==nil{t.Fatal("unexpected valid truncated control file")}
}
