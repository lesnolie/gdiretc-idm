package main

import (
 "encoding/binary"
 "errors"
 "fmt"
 "math"
 "os"
)

// aria2 writes the .aria2 control file atomically every few seconds.
// For parallel range downloads the size of the output file is NOT a
// measure of bytes actually received: writes can be sparse or out of order.
// This reader counts only completed aria2 pieces, so the visual progress
// is conservative until the next checkpoint. It never modifies control data.
func readAria2Progress(output string)(int64,int64,error){
 data,err:=os.ReadFile(output+".aria2")
 if err!=nil{return 0,0,err}
 if len(data)<34{return 0,0,errors.New("short aria2 control file")}
 if binary.BigEndian.Uint16(data[0:2])!=1{return 0,0,errors.New("unsupported aria2 control version")}
 // This app downloads only individual HTTP files, not BitTorrent.
 if binary.BigEndian.Uint32(data[2:6])!=0{return 0,0,errors.New("unexpected torrent control data")}
 hashBytes:=binary.BigEndian.Uint32(data[6:10])
 if hashBytes!=0{return 0,0,errors.New("unexpected torrent info hash")}
 pieceSize:=uint64(binary.BigEndian.Uint32(data[10:14]))
 total:=binary.BigEndian.Uint64(data[14:22])
 fieldLen:=uint64(binary.BigEndian.Uint32(data[30:34]))
 if pieceSize==0||total==0||total>math.MaxInt64{
  return 0,0,errors.New("invalid aria2 piece or file size")
 }
 pieces:=total/pieceSize
 if total%pieceSize!=0{pieces++}
 required:=pieces/8
 if pieces%8!=0{required++}
 if fieldLen<required || fieldLen>4<<20 {
  return 0,0,fmt.Errorf("unexpected aria2 bitfield length %d",fieldLen)
 }
 if uint64(len(data))<34+fieldLen{return 0,0,errors.New("incomplete aria2 bitfield")}
 field:=data[34:34+int(fieldLen)]
 var completed uint64
 for n,b:=range field{
  if b==0{continue}
  for bit:=0;bit<8;bit++{
   if b&(1<<uint(7-bit))==0{continue}
   index:=uint64(n*8+bit)
   if index>=pieces{continue}
   size:=pieceSize
   if index==pieces-1{
    if tail:=total-index*pieceSize;tail<size{size=tail}
   }
   if completed>math.MaxInt64-size{return 0,0,errors.New("invalid aria2 completed size")}
   completed+=size
  }
 }
 // Reject inconsistent control information rather than jumping the bar.
 if completed>total{return 0,0,errors.New("aria2 completed length exceeds total")}
 return int64(completed),int64(total),nil
}
