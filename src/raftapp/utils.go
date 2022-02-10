package raftapp

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
)

func PrettyPrint(x interface{}) string {
  s,_:=json.Marshal(x)
  return string(s)
}

func ZipData(snapshot Snapshot) Snapshot {
  buf:=new(bytes.Buffer)
  zw:=gzip.NewWriter(buf)
  _,err:=zw.Write(snapshot)
  if err!=nil {
    panic(err)
  }
  if err := zw.Close(); err != nil {
    panic(err)
  }
  return buf.Bytes()
}

func UnzipData(snapshot Snapshot) Snapshot {
  zr,err:=gzip.NewReader(bytes.NewBuffer(snapshot))
  if err!=nil {
    panic(err)
  }
  buf:=new(bytes.Buffer)
  io.Copy(buf, zr)
  if err:=zr.Close(); err!=nil {
    panic(err)
  }
  return buf.Bytes()
}
