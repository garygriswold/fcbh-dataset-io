package output

import (
	"encoding/csv"
	"os"
	"reflect"
	"strconv"
)

func WriteCSV(structs []any, meta []Meta) string {
	file, err := os.CreateTemp(os.Getenv(`FCBH_DATASET_TMP`), "csv")
	if err != nil {
		panic(err)
	}
	writer := csv.NewWriter(file)
	var header []string
	for _, mt := range meta {
		if mt.Cols == 1 {
			header = append(header, mt.Tag)
		} else {
			for i := 0; i < mt.Cols; i++ {
				header = append(header, mt.Tag+strconv.Itoa(i))
			}
		}
	}
	_ = writer.Write(header)
	for _, scr := range structs {
		str := reflect.ValueOf(scr)
		var line = make([]string, len(header))
		for _, mt := range meta {
			data := str.Field(mt.Index)
			if data.Kind() == reflect.Slice {
				for i := 0; i < data.Len(); i++ {
					if data.Index(i).Kind() == reflect.Slice {
						for j := 0; j < data.Index(i).Len(); j++ {
							line[mt.CSVPos+j] = ToString(data.Index(i).Index(j))
						}
						_ = writer.Write(line)
						line = make([]string, len(header))
					} else {
						line[mt.CSVPos+i] = ToString(data.Index(i))
					}
				}
			} else {
				line[mt.CSVPos] = ToString(data)
			}
		}
		_ = writer.Write(line)
		line = make([]string, len(header))
	}
	writer.Flush()
	err = writer.Error()
	if err != nil {
		panic(err)
	}
	_ = file.Close()
	return file.Name()
}

// ToString converts scalar values to string.  It does not convert the following kind.
// reflect.Uintptr, reflect.Complex64, reflect.Complex128, reflect.Struct, reflect.Array,
// reflect.Slice, reflect.Chan, reflect.Func, reflect.Interface, reflect.Invalid
func ToString(value reflect.Value) string {
	var result string
	switch value.Kind() {
	case reflect.Bool:
		result = strconv.FormatBool(value.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		result = strconv.FormatInt(value.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		result = strconv.FormatUint(value.Uint(), 10)
	case reflect.Float32:
		result = strconv.FormatFloat(value.Float(), 'f', -1, 32)
	case reflect.Float64:
		result = strconv.FormatFloat(value.Float(), 'f', -1, 32)
	case reflect.String:
		result = value.String()
	default:
		panic(`output.ToString() cannot convert value of type` + value.Type().String())
	}
	return result
}
