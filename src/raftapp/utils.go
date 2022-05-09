package raftapp

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"

	"6.824/labgob"
)

func PrettyPrint(x interface{}) string {
	// s,err:=json.Marshal(x)
	// if err!=nil {
	//   return fmt.Sprintf("%+v", x)
	// } else {
	//   return string(s)
	// }
	return fmt.Sprintf("%+v", x)
}

func ZipData(snapshot Snapshot) Snapshot {
	buf := new(bytes.Buffer)
	zw := gzip.NewWriter(buf)
	_, err := zw.Write(snapshot)
	if err != nil {
		panic(err)
	}
	if err := zw.Close(); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func UnzipData(snapshot Snapshot) Snapshot {
	zr, err := gzip.NewReader(bytes.NewBuffer(snapshot))
	if err != nil {
		panic(err)
	}
	buf := new(bytes.Buffer)
	io.Copy(buf, zr)
	if err := zr.Close(); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func ValueToSnapshot(x interface{}) Snapshot {
	buffer := new(bytes.Buffer)
	encoder := labgob.NewEncoder(buffer)
	err := encoder.Encode(x)
	if err != nil {
		panic(err)
	}
	return buffer.Bytes()
}

func SnapshotToValue(b Snapshot, x interface{}) {
	decoder := labgob.NewDecoder(bytes.NewBuffer(b))
	err := decoder.Decode(x)
	if err != nil {
		panic(err)
	}
}

func DeepCopy(src interface{}, dst interface{}) {
	SnapshotToValue(ValueToSnapshot(src), dst)
}
