package main
import ("testing")
func TestMetadataParsing(t *testing.T){
 links:="198.51.100.10\tOK\t91 ms\n203.0.113.15\tOK\t82 ms (SNI g.cn)\n1.1.1.1\tunavailable\ttimeout\n"
 eps:=parseSavedEndpoints(links)
 if len(eps)!=2 || eps[0].ip!="198.51.100.10"||eps[1].sni!="g.cn"{t.Fatalf("unexpected endpoints: %#v",eps)}
 if got:=nameFromDisposition(`attachment; filename*=UTF-8''demo%20test.zip`);got!="demo test.zip"{t.Errorf("filename*: %q",got)}
 if got:=nameFromHTML(`<!doctype html><meta property="og:title" content="hello &amp; world.zip">`);got!="hello & world.zip"{t.Errorf("og:title: %q",got)}
 if got:=safeName("../../etc/passwd");got!="passwd"{t.Errorf("unsafe path: %q",got)}
 if allowedMetadataHost("drive.google.com.evil.org"){t.Fatal("untrusted host accepted")}
}

func TestCopyNameNormalization(t *testing.T){
 tests:=[]struct{name,typ,want string}{
  {"旅行.mp4 的副本","application/octet-stream","旅行 的副本.mp4"},
  {"视频.mp4 的副本 (2)","application/octet-stream","视频 的副本 (2).mp4"},
  {"movie.mp4 - Copy","application/octet-stream","movie - Copy.mp4"},
  {"旅行 的副本","video/mp4","旅行 的副本.mp4"},
  {"test","video/quicktime; charset=binary","test.mov"},
  {"already.mp4","video/mp4","already.mp4"},
  {"archive.zip","application/octet-stream","archive.zip"},
  {"untitled","application/octet-stream","untitled"},
 }
 for _,tc:=range tests{
  if got:=normalizeDownloadName(tc.name,tc.typ);got!=tc.want{
   t.Errorf("normalizeDownloadName(%q,%q)=%q, want %q",tc.name,tc.typ,got,tc.want)
  }
 }
}
func TestEndpointLatencyParsing(t *testing.T){
 raw:="142.250.2.3\tOK\t24 ms\n142.250.3.4\tOK\t75 ms (SNI g.cn)\n8.8.8.8\tOK\t10 ms [not Google; ignored]\n"
 got:=parseSavedEndpoints(raw)
 if len(got)!=2||got[0].latencyMS!=24||got[1].latencyMS!=75||got[1].sni!="g.cn"{
  t.Fatalf("parseSavedEndpoints: %#v",got)
 }
}
