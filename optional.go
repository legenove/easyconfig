package easyconfig

import "github.com/legenove/easyconfig/ifacer"

func OptOnChangeFunc(f ifacer.ChangeFunc) ifacer.OptionFunc {
	return func(iv ifacer.Configer) {
		iv.SetOnChangeFunc(f)
	}
}

func OptOnRemoveFunc(f ifacer.ChangeFunc) ifacer.OptionFunc {
	return func(iv ifacer.Configer) {
		iv.SetOnRemoveFunc(f)
	}
}

func OptConfType(t string) ifacer.OptionFunc {
	return func(iv ifacer.Configer) {
		iv.SetConfType(t)
	}
}
