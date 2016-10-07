package storm

import "reflect"

// Update a structure
func (n *node) Update(data interface{}) error {
	return n.update(data, func(ref *reflect.Value, current *reflect.Value, cfg *structConfig) error {
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
				idxInfo, ok := cfg.Fields[ref.Type().Field(i).Name]
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
	return n.update(data, func(ref *reflect.Value, current *reflect.Value, cfg *structConfig) error {
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
		idxInfo, ok := cfg.Fields[fieldName]
		if ok {
			idxInfo.Value = &f
			idxInfo.IsZero = isZero(idxInfo.Value)
		}
		return nil
	})
}

func (n *node) update(data interface{}, fn func(*reflect.Value, *reflect.Value, *structConfig) error) error {
	ref := reflect.ValueOf(data)

	if !ref.IsValid() || ref.Kind() != reflect.Ptr || ref.Elem().Kind() != reflect.Struct {
		return ErrStructPtrNeeded
	}

	cfg, err := extract(&ref)
	if err != nil {
		return err
	}

	if cfg.ID.IsZero {
		return ErrNoID
	}

	id, err := toBytes(cfg.ID.Value.Interface(), n.s.codec)
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
	err = ntx.One(cfg.ID.Name, cfg.ID.Value.Interface(), current.Interface())
	if err != nil {
		return err
	}

	ref = ref.Elem()
	cref := current.Elem()
	err = fn(&ref, &cref, cfg)
	if err != nil {
		return err
	}

	raw, err := ntx.s.codec.Marshal(current.Interface())
	if err != nil {
		return err
	}

	err = ntx.save(ntx.tx, cfg, id, raw, nil)
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
