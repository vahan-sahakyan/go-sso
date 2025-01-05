package utils

import (
  "encoding/json"
  "fmt"
)

func StringifyStruct(v interface{}) string {
  result, _ := json.MarshalIndent(v, "", "    ")
  return fmt.Sprintf("%T", v) + string(result)
}
