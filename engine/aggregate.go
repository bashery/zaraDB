package engine

import (
	"fmt"
	srt "sort"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type Aggregate struct{}

func order(data []string, params gjson.Result) []string {

	var tmplist []gjson.Result
	for _, v := range data {
		tmplist = append(tmplist, gjson.Parse(v))
	}

	var fields []string
	var asc []bool

	params.ForEach(func(f, v gjson.Result) bool {
		fields = append(fields, f.Str)
		if v.Int() == 1 {
			asc = append(asc, true)
			return true
		}
		asc = append(asc, false)
		return true
	})

	// types : 3 is string, 2 is Num, 5 is json or array.

	srt.Slice(tmplist, func(i, j int) bool {
		for index, field := range fields {
			valueI := tmplist[i].Get(field)
			valueJ := tmplist[j].Get(field)

			if asc[index] {
				if valueI.Value() != valueJ.Value() {
					switch valueI.Type {
					case 3:
						return valueI.Str < valueJ.Str

					case 2:
						return valueI.Num < valueJ.Num

					default:
						fmt.Println("unsupported type")
						return true
					}
				}
			}

			if valueI.Value() != valueJ.Value() {
				switch valueI.Type {
				case 3:
					return valueI.Str > valueJ.Str

				case 2:
					return valueI.Num > valueJ.Num

				default:
					fmt.Println("unsupported type")
					return true
				}
			}
		}
		return false
	})
	for k, v := range tmplist {
		data[k] = v.Raw
	}

	return data
}

// not implemented yet
func (ag Aggregate) aggrigate(query gjson.Result) string {
	data, err := db.getData(query)
	if err != nil {
		return err.Error()
	}
	if len(data) == 0 {
		return "[]"
	}

	group := query.Get("group")
	if _id := group.Get("_id"); !_id.Exists() {
		return `{"code":0, "status":"a group specification must include an _id"}`
	}

	mapData := map[string]string{}
	message := ""

	_id := group.Get("_id").Str

	// TODO parse data and exclude invalide objects
	if gjson.Get(data[0], _id).Str == "" {
		return "field '" + _id + "' is not exists"
	}

	group.ForEach(func(key, val gjson.Result) bool {
		if key.Str != "_id" && val.Type != 5 {
			message = "The field '" + key.Str + "' must be an accumulator object!"
			return false
		}

		switch val.Type {
		case 3:
			for _, obj := range data {
				field := gjson.Get(obj, val.String()).String()
				json, _ := sjson.Set("", key.String(), field)
				mapData[field] = json
			}

		case 5:
			// opr operation, fld field name
			val.ForEach(func(opr, fld gjson.Result) bool {

				switch opr.Str {
				case "$count":
					counted := ag.count(_id, data)
					for _id, count := range counted {
						mapData[_id], _ = sjson.Set(mapData[_id], key.Str, count)
					}

				case "$max":
					maxs, err := max(_id, fld, data)
					if err != nil {
						message = err.Error()
						return false
					}
					for _id, max := range maxs {
						mapData[_id], _ = sjson.Set(mapData[_id], key.Str, max)
					}

				case "$min":
					mins, err := min(_id, fld, data)
					if err != nil {
						message = err.Error()
						return false
					}

					for _id, min := range mins {
						mapData[_id], _ = sjson.Set(mapData[_id], key.Str, min)
					}

				case "$sum":
					sums, err := ag.sum(_id, fld, data)
					if err != nil {
						message = err.Error()
						return false
					}

					for _id, sumd := range sums {
						mapData[_id], _ = sjson.Set(mapData[_id], key.Str, sumd)
					}

				case "$avg":
					avrs, err := average(_id, fld, data)
					if err != nil {
						message = err.Error()
						return false
					}
					for _id, avr := range avrs {
						mapData[_id], _ = sjson.Set(mapData[_id], key.Str, avr)
					}

				default:
					message = "unknown '" + opr.Str + "' aggrigate operator !"
					return false
				}

				return true
			})

		default:
			message = "opss! something wrrong!"
			return false
		}

		return true
	})

	if message != "" {
		return message
	}

	filter := query.Get(gmatch)
	limit := query.Get(glimit).Int()
	skip := query.Get(gskip).Int()

	if limit == 0 {
		// what is default setting ?
		limit = 1000
	}

	listdata := []string{}
	for _, val := range mapData {
		if ok, _ := match(filter, val); ok {
			listdata = append(listdata, val)
		}
	}

	// TODO sort listdata here
	srt := query.Get("gsort")
	if srt.Exists() {
		listdata = order(listdata, srt)
	}

	result := "["
	for _, val := range listdata {

		if limit == 0 {
			break
		}
		if skip != 0 {
			skip--
			continue
		}

		listdata = append(listdata, val)
		result += val + ","
		limit--
	}

	ln := len(result)
	if ln == 1 {
		return "[]"
	}

	return result[:ln-1] + "]"

}

func (ag Aggregate) sum(_id string, field gjson.Result, records []string) (mp map[string]float64, err error) {
	mp = map[string]float64{}

	switch field.Type {
	case 3:
		for _, record := range records {
			id := gjson.Get(record, _id).Str
			val := gjson.Get(record, field.Str).Num

			mp[id] += val
		}

	case 5:

		field.ForEach(func(op, args gjson.Result) bool {
			switch op.Str {
			case "$multiply":

				for _, record := range records {
					id := gjson.Get(record, _id).Str
					arg1 := gjson.Get(record, args.Get("0").Str).Num
					arg2 := gjson.Get(record, args.Get("1").Str).Num

					mp[id] += arg1 * arg2
				}
			case "$add":

				for _, record := range records {
					fmt.Println(record)

					id := gjson.Get(record, _id).Str
					arg1 := gjson.Get(record, args.Get("0").Str).Num
					arg2 := gjson.Get(record, args.Get("1").Str).Num

					mp[id] += arg1 + arg2
				}

			case "$sub":

				for _, record := range records {

					id := gjson.Get(record, _id).Str
					arg1 := gjson.Get(record, args.Get("0").Str).Num
					arg2 := gjson.Get(record, args.Get("1").Str).Num

					mp[id] += arg1 - arg2
				}

			case "$div":

				for _, record := range records {
					fmt.Println(record)

					id := gjson.Get(record, _id).Str
					arg1 := gjson.Get(record, args.Get("0").Str).Num
					arg2 := gjson.Get(record, args.Get("1").Str).Num

					mp[id] += arg1 / arg2
				}

			default:
				err = fmt.Errorf("unknown %s operator", op)
			}

			return false
		})
		if err != nil {
			return nil, err
		}

		fmt.Println(" we are here ")
	}

	return mp, nil
}

// gets mines vlaues per _id (e.g. name)
func min(_id string, field gjson.Result, records []string) (mp map[string]float64, err error) {
	min := float64(9223372036854775807) // we need max float
	mp = map[string]float64{}

	// init mp by max float val
	for _, record := range records {
		id := gjson.Get(record, _id).Str
		mp[id] = min
	}

	switch field.Type {
	case 3:
		for _, record := range records {
			id := gjson.Get(record, _id).Str
			val := gjson.Get(record, field.Str).Num
			if mp[id] > val {
				mp[id] = val
			}
		}

	case 5:

		field.ForEach(func(op, args gjson.Result) bool {
			switch op.Str {
			case "$multiply":
				for _, record := range records {

					id := gjson.Get(record, _id).Str
					arg1 := gjson.Get(record, args.Get("0").Str).Num
					arg2 := gjson.Get(record, args.Get("1").Str).Num

					val := arg1 * arg2
					if mp[id] > val {
						mp[id] = val
					}
				}
			case "$add":

				for _, record := range records {
					fmt.Println(record)

					id := gjson.Get(record, _id).Str
					arg1 := gjson.Get(record, args.Get("0").Str).Num
					arg2 := gjson.Get(record, args.Get("1").Str).Num

					val := arg1 + arg2
					if mp[id] > val {
						mp[id] = val
					}
				}

			case "$sub":

				for _, record := range records {
					fmt.Println(record)

					id := gjson.Get(record, _id).Str
					arg1 := gjson.Get(record, args.Get("0").Str).Num
					arg2 := gjson.Get(record, args.Get("1").Str).Num

					val := arg1 - arg2
					if mp[id] > val {
						mp[id] = val
					}
				}

			case "$div":

				for _, record := range records {
					fmt.Println(record)

					id := gjson.Get(record, _id).Str
					arg1 := gjson.Get(record, args.Get("0").Str).Num
					arg2 := gjson.Get(record, args.Get("1").Str).Num
					val := arg1 / arg2

					if mp[id] > val {
						mp[id] = val
					}
				}

			default:
				err = fmt.Errorf("unknown %s operator", op)
			}

			return false
		})
		if err != nil {
			return nil, err
		}
	}

	return mp, nil
}

// max gets vlaues per _id (e.g. name)
func max(_id string, field gjson.Result, records []string) (mp map[string]float64, err error) {
	max := float64(-9223372036854775808) // we need min floa
	mp = map[string]float64{}

	// init mp by mines float val
	for _, record := range records {
		id := gjson.Get(record, _id).Str
		mp[id] = max
	}

	switch field.Type {
	case 3:
		for _, record := range records {
			id := gjson.Get(record, _id).Str
			val := gjson.Get(record, field.Str).Num
			if mp[id] < val {
				mp[id] = val
			}
		}

	case 5:

		field.ForEach(func(op, args gjson.Result) bool {
			switch op.Str {
			case "$multiply":

				for _, record := range records {

					id := gjson.Get(record, _id).Str
					arg1 := gjson.Get(record, args.Get("0").Str).Num
					arg2 := gjson.Get(record, args.Get("1").Str).Num

					val := arg1 * arg2
					if mp[id] < val {
						mp[id] = val
					}
				}
			case "$add":

				for _, record := range records {
					fmt.Println(record)

					id := gjson.Get(record, _id).Str
					arg1 := gjson.Get(record, args.Get("0").Str).Num
					arg2 := gjson.Get(record, args.Get("1").Str).Num

					val := arg1 + arg2
					if mp[id] < val {
						mp[id] = val
					}
				}

			case "$sub":

				for _, record := range records {
					fmt.Println(record)

					id := gjson.Get(record, _id).Str
					arg1 := gjson.Get(record, args.Get("0").Str).Num
					arg2 := gjson.Get(record, args.Get("1").Str).Num

					val := arg1 - arg2
					if mp[id] < val {
						mp[id] = val
					}
				}

			case "$div":

				for _, record := range records {
					fmt.Println(record)

					id := gjson.Get(record, _id).Str
					arg1 := gjson.Get(record, args.Get("0").Str).Num
					arg2 := gjson.Get(record, args.Get("1").Str).Num
					val := arg1 / arg2

					if mp[id] < val {
						mp[id] = val
					}
				}

			default:
				err = fmt.Errorf("unknown %s operator", op)
			}

			return false
		})

		if err != nil {
			return nil, err
		}
	}

	return mp, nil
}

// not implemented yet
func average(_id string, field gjson.Result, records []string) (mp map[string]float64, err error) {
	mp = map[string]float64{}

	fieldCount := make(map[string]float64)

	switch field.Type {
	case 3:
		for _, record := range records {
			id := gjson.Get(record, _id).Str
			val := gjson.Get(record, field.Str).Num

			//if val.Exists() {}
			fieldCount[id]++
			mp[id] += val //.Num
		}

	case 5:

		field.ForEach(func(op, args gjson.Result) bool {
			switch op.Str {
			case "$multiply":

				for _, record := range records {

					id := gjson.Get(record, _id).Str
					arg1 := gjson.Get(record, args.Get("0").Str).Num
					arg2 := gjson.Get(record, args.Get("1").Str).Num

					fieldCount[id]++
					mp[id] += arg1 * arg2
				}

			case "$add":

				for _, record := range records {
					fmt.Println(record)

					id := gjson.Get(record, _id).Str
					arg1 := gjson.Get(record, args.Get("0").Str).Num
					arg2 := gjson.Get(record, args.Get("1").Str).Num

					fieldCount[id]++
					mp[id] += arg1 + arg2
				}

			case "$sub":

				for _, record := range records {
					fmt.Println(record)

					id := gjson.Get(record, _id).Str
					arg1 := gjson.Get(record, args.Get("0").Str).Num
					arg2 := gjson.Get(record, args.Get("1").Str).Num

					val := arg1 - arg2
					mp[id] += val //.Num

					fieldCount[id]++

				}

			case "$div":

				for _, record := range records {
					fmt.Println(record)

					id := gjson.Get(record, _id).Str
					arg1 := gjson.Get(record, args.Get("0").Str).Num
					arg2 := gjson.Get(record, args.Get("1").Str).Num
					val := arg1 / arg2

					mp[id] += val //.Num
					fieldCount[id]++

				}

			default:
				err = fmt.Errorf("unknown %s operator", op)
			}

			return false
		})

		if err != nil {
			return nil, err
		}
	}

	for fld, count := range fieldCount {
		mp[fld] = mp[fld] / count
	}

	return mp, nil
} // average

// count counts entities in collection by provid field
func (ag Aggregate) count(field string, records []string) (mp map[string]int) {

	mp = map[string]int{}
	for _, record := range records {
		mp[gjson.Get(record, field).Str]++
	}

	return mp
}

// count counts entities in collection by provid field
func countDoc(records []string) string {
	return strconv.Itoa(len(records))
}

// count counts entities in collection by provid field
func countField(field string, records []string) string {
	res := 0
	for _, record := range records {
		if gjson.Get(record, field).Exists() {
			res++
		}
	}

	return strconv.Itoa(res)
}

// reKey renames json feild
func reKey(oldkey, newkey, json string) string {

	list := []string{}

	gjson.Parse(json).ForEach(func(key, value gjson.Result) bool {

		if key.String() == oldkey {
			list = append(list, newkey+siparator+value.String())

		} else {
			list = append(list, key.String()+siparator+value.String())
		}

		return true
	})

	slice := []string{}

	res := "{"
	for _, v := range list {
		if string(v[len(v)-1]) == "}" {
			continue
		}

		slice = strings.Split(v, siparator)

		if gjson.Parse(slice[1]).Type == 2 {
			res += `"` + slice[0] + `":` + slice[1] + `,`
			continue
		}
		res += `"` + slice[0] + `":"` + slice[1] + `",`
	}

	return res[:len(res)-1] + `}`
}

// reFields remove or rename fields in result data
func reFields(data []string, fields gjson.Result) []string {

	newKey := []string{}
	oldKey := []string{}

	toRemove := make([]string, 0)

	for k, v := range fields.Map() {
		if v.String() == "0" {
			toRemove = append(toRemove, k)
		} else {
			newKey = append(newKey, v.String())
			oldKey = append(oldKey, k)
		}
	}

	//remove fields
	for i := 0; i < len(data); i++ {
		for _, k := range toRemove {
			data[i], _ = sjson.Delete(data[i], k)
		}
	}

	//reName fields
	for i := 0; i < len(data); i++ {
		for j := 0; j < len(newKey); j++ {
			data[i] = reKey(oldKey[j], newKey[j], data[i])
		}
	}

	return data
}

// end
