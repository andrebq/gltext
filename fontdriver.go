// Copyright 2012 The go-gl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gltext

import (
	"errors"
	"image"
	"io"
	"sync"
)

// Driver declares the signature required to enable easy replacement of
// the text rendering engine.
type Driver func(r io.Reader, scale int32, low, high rune, dir Direction) (*FontConfig, *image.RGBA, error)

type driverManager struct {
	sync.RWMutex
	drivers       map[string]Driver
	currentDriver Driver
}

var (
	dm = driverManager{
		drivers:       make(map[string]Driver),
		currentDriver: nil,
	}
)

// Register the given Driver under the given name.
//
// If a name is already used by other driver, this
// will panic
func RegisterDriver(name string, driver Driver) {
	dm.Lock()
	defer dm.Unlock()
	if _, has := dm.drivers[name]; has {
		panic(name + " already used")
	}
	dm.drivers[name] = driver
}

// Select the driver that should be used to render new glyphs,
// old configurations aren't affected.
func UseDriver(name string) error {
	dm.Lock()
	defer dm.Unlock()
	if d, has := dm.drivers[name]; has {
		dm.currentDriver = d
		return nil
	}
	return errors.New(name + " isn't registred")
}

// MustUseDriver checks if the driver is valid and use it,
// otherwise this panic
func MustUseDriver(name string) {
	err := UseDriver(name)
	if err != nil {
		panic(err)
	}
}

func LoadTruetype(r io.Reader, scale int32, low, high rune, dir Direction) (*Font, error) {
	dm.Lock()
	cur := dm.currentDriver
	dm.Unlock()
	if cur == nil {
		return nil, errors.New("no driver is set")
	}
	fc, img, err := cur(r, scale, low, high, dir)
	if err != nil {
		return nil, err
	}
	return loadFont(img, fc)
}
