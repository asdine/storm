package storm

import "reflect"

// Update a structure
func (n *node) Update(data interface{}) error {
	return n.update(data, func(ref *reflect.Value, current *reflect.Value, info *modelInfo) error {
		numfield := ref.NumField()
		for i := 0; i < numfield; i++ {
			f := ref.Field(i)
			if ref.Type().Field(i).PkgPath != "" {
				continue
			}
			zero := reflect.Zero(f.Type()).Interface()
			actual := f.Interface()
			if !reflect.DeepEqual(actual, zero) {
				cf := current.Field(i)
				cf.Set(f)
				idxInfo, ok := info.Indexes[ref.Type().Field(i).Name]
				if ok {
					idxInfo.Value = &cf
				}
			}
		}
		return nil
	})
}

// UpdateField updates a single field
func (n *node) UpdateField(data interface{}, fieldName string, value interface{}) error {
	return n.update(data, func(ref *reflect.Value, current *reflect.Value, info *modelInfo) error {
		f := current.FieldByName(fieldName)
		if !f.IsValid() {
			return ErrNotFound
		}
		tf, _ := current.Type().FieldByName(fieldName)
		if tf.PkgPath != "" {
			return ErrNotFound
		}
		v := reflect.ValueOf(value)
		if v.Kind() != f.Kind() {
			return ErrIncompatibleValue
		}
		f.Set(v)
		idxInfo, ok := info.Indexes[fieldName]
		if ok {
			idxInfo.Value = &f
			idxInfo.IsZero = idxInfo.isZero()
		}
		return nil
	})
}

func (n *node) update(data interface{}, fn func(*reflect.Value, *reflect.Value, *modelInfo) error) error {
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

	id, err := toBytes(info.ID.Value.Interface(), n.s.codec)
	if err != nil {
		return err
	}

	current := reflect.New(reflect.Indirect(ref).Type())

	tx, err := n.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	ntx := tx.(*node)
	err = ntx.One(info.ID.FieldName, info.ID.Value.Interface(), current.Interface())
	if err != nil {
		return err
	}

	ref = ref.Elem()
	cref := current.Elem()
	err = fn(&ref, &cref, info)
	if err != nil {
		return err
	}

	raw, err := ntx.s.codec.Encode(current.Interface())
	if err != nil {
		return err
	}

	err = ntx.save(ntx.tx, info, id, raw, nil)
	if err != nil {
		return err
	}

	return ntx.Commit()
}

// Update a structure
func (s *DB) Update(data interface{}) error {
	return s.root.Update(data)
}

// UpdateField updates a single field
func (s *DB) UpdateField(data interface{}, fieldName string, value interface{}) error {
	return s.root.UpdateField(data, fieldName, value)
}
