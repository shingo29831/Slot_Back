package main
import "reflect"

type Array struct{
	data []interface{}
	size int 
}

func (array *Array) append(data any){
	array.data = append(array.data, data)
	array.size++
}

func initArray()(*Array){
	var array Array
	array.size = 0
	return &array
}

func (array * Array)del_data(data any)(bool){
	for i := 0; i < array.size; i++ {
		if reflect.DeepEqual(array.data[i],data){
			tmp := array.data[0:i]
			array.data = append(array.data[(i + 1):], tmp)
			return true
		}
	}
	return false
}
