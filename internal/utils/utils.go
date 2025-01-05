package utils

import "encoding/json"

func StringifyStruct(v interface{}, structName string) string {
  res, _ := json.MarshalIndent(v, "", "    ")
  return structName + string(res)
}
