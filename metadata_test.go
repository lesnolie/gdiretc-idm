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
