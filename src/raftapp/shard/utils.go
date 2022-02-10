package shard

import (
	"sort"
	"strings"
)

func CmpStringSlice(a []string, b []string) int {
  sort.Strings(a)
  sort.Strings(b)
  for i:=0;i<len(a)&&i<len(b);i++ {
    t:=strings.Compare(a[i],b[i])
    if t==0 {
      continue
    } else {
      return t
    }
  }
  if len(a) == len(b) {
    return 0
  } else if len(a) < len(b) {
    return -1
  } else {
    return 1
  }
}

