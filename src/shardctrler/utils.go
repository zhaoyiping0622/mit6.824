package shardctrler

func EqualStringSlice(a []string, b []string) bool {
  if len(a)==len(b){
    for i:=range a {
      if a[i] != b[i] {
        return false;
      }
    }
    return true;
  }
  return false
}
