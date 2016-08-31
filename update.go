package storm

import "reflect"

// Update a structure
func (n *node) Update(data interface{}) error {
	ref := reflect.ValueOf(data)

	if !ref.IsValid() || ref.Kind() != reflect.Ptr || ref.Elem().Kind() != reflect.Struct {
		return ErrStructPtrNeeded
	}

	info, err := extract(&ref)
	if err != nil {
		return err
	}

	if info.ID.IsZero {
		return ErrNoID
	}

	id, err := toBytes(info.ID.Value.Interface(), n.s.Codec)
	if err != nil {
		return err
	}

	current := reflect.New(reflect.Indirect(ref).Type())

	tx, err := n.Begin(true)
	if err != nil {
		return err
	}

	ntx := tx.(*node)

	err = ntx.One(info.ID.FieldName, info.ID.Value.Interface(), current.Interface())
	if err != nil {
		ntx.Rollback()
		return err
	}

	ref = ref.Elem()
	numfield := ref.NumField()
	for i := 0; i < numfield; i++ {
		f := ref.Field(i)
		zero := reflect.Zero(f.Type()).Interface()
		if ref.Type().Field(i).PkgPath != "" {
			continue
		}
		actual := f.Interface()
		if !reflect.DeepEqual(actual, zero) {
			cf := current.Elem().Field(i)
			cf.Set(f)
			idxInfo, ok := info.Indexes[ref.Type().Field(i).Name]
			if ok {
				idxInfo.Value = &cf
			}
		}
	}

	raw, err := ntx.s.Codec.Encode(current.Interface())
	if err != nil {
		ntx.Rollback()
		return err
	}

	err = ntx.save(ntx.tx, info, id, raw, nil)
	if err != nil {
		ntx.Rollback()
		return err
	}

	return ntx.Commit()
}

// Update a structure
func (s *DB) Update(data interface{}) error {
	return s.root.Update(data)
}
