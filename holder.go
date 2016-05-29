package summer

import (
	"reflect"
	"strings"
	"qiniupkg.com/x/log.v7"
)
// a holder that can hold stone
type Holder struct {
	Stone        Stone
	Class        reflect.Type
	PointerClass reflect.Type
	Value        reflect.Value
	Basket       *basket
	Dependents   []*Holder
}

func newHolder(stone Stone, basket *basket) *Holder {
	return &Holder{
		Stone:stone,
		Class: reflect.TypeOf(stone).Elem(),
		PointerClass: reflect.TypeOf(stone),
		Value: reflect.ValueOf(stone).Elem(),
		Basket: basket,
		Dependents: []*Holder{},
	}
}
func (this *Holder) ResolveDirectlyDependents() {
	num := this.Value.NumField() - 1
	for ; num >= 0; num-- {
		this.SetDirectDependValue(this.Value.Field(num), this.Class.Field(num))
	}
}
// in this step we try to find the stone which the field need
func (this *Holder) SetDirectDependValue(fieldValue reflect.Value, fieldInfo reflect.StructField) {
	// get the field's tag which belongs to summer
	tag := fieldInfo.Tag.Get("sm")
	if tag == "" {
		return
	}
	log.Println("[build Field]", this.Class.Name(), fieldInfo.Name, fieldInfo.Type.Name(), fieldInfo.Tag, tag)
	// convert text to summer tag option
	tagOption := buildTagOptions(tag)
	// if the field not a straight depend
	if !tagOption.depend {
		// may be the plugin will help it
		this.Basket.PutDelayField(&DelayField{fieldValue, fieldInfo, tagOption, this})
		log.Println(this.Class.Name(), " the field [", fieldInfo.Name, "] will be delay. ", tagOption)
		return
	}
	// get stone's name which the field wanted
	var name string
	if tagOption.auto {
		name = fieldInfo.Name
		name = strings.ToLower(name[:1]) + name[1:]
	} else {
		name = tagOption.name
	}
	// get the field type
	fieldType := fieldValue.Type()
	// find the needed stone holder from basket
	hd := this.Basket.findHolder(name, fieldType)
	// if holder not found
	if hd == nil {
		// maybe the name is wrong,we suggest the type'name is the stone's name
		if fieldType.Kind() == reflect.Ptr {
			name = fieldType.Elem().Name()
		} else {
			name = fieldType.Name()
		}
		name = strings.ToLower(name[:1]) + name[1:]
		hd = this.Basket.findHolder(name, fieldType)
		if hd == nil {
			// we don't know what happened ,maybe you forget put the stone into the basket
			// so just panic
			panic(CannotResolveDependencyErr)
		}
	}
	// don't forget to record the dependency of the stone we need
	this.Dependents = append(this.Dependents, hd)
	fieldValue.Set(reflect.ValueOf(hd.Stone))
	log.Println(this.Class.Name(), " depend on ", hd.Class.Name())
}
func (this *Holder)init(holders map[*Holder]bool) {
	if stone, ok := this.Stone.(Init); ok {
		if holders[this] {
			return
		}
		holders[this] = true
		for _, v := range this.Dependents {
			v.ready(holders)
		}
		stone.Init()
	}
}
func (this *Holder)ready(holders map[*Holder]bool) {
	if stone, ok := this.Stone.(Ready); ok {
		if holders[this] {
			return
		}
		holders[this] = true
		for _, v := range this.Dependents {
			v.ready(holders)
		}
		stone.Ready()
	}
}
func (this *Holder)destroy(holders map[*Holder]bool) {
	if stone, ok := this.Stone.(Destroy); ok {
		if holders[this] {
			return
		}
		holders[this] = true
		for _, v := range this.Dependents {
			v.destroy(holders)
		}
		stone.Destroy()
	}
}