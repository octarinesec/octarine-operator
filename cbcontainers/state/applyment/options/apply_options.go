package options

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const (
	DefaultCreateOnlyValue = false
)

type OwnerSetter func(controlledResource metav1.Object) error

type ApplyOptions struct {
	//When set to true, The k8s object will not be modified if it already exists
	//Default set to false
	createOnly *bool

	//The callback that sets the owner of the k8s object
	//Default set to nil
	setOwner OwnerSetter
}

func MergeApplyOptions(options ...*ApplyOptions) *ApplyOptions {
	mergedApplyOptions := NewApplyOptions()

	for _, singleApplyOptions := range options {

		if singleApplyOptions.createOnly != nil {
			mergedApplyOptions.createOnly = singleApplyOptions.createOnly
		}

		if singleApplyOptions.setOwner != nil {
			mergedApplyOptions.setOwner = singleApplyOptions.setOwner
		}
	}

	return mergedApplyOptions
}

func NewApplyOptions() *ApplyOptions {
	return &ApplyOptions{
		createOnly: nil,
		setOwner:   nil,
	}
}

func (options *ApplyOptions) CreateOnly() bool {
	if options.createOnly != nil {
		return *options.createOnly
	}

	return DefaultCreateOnlyValue
}

func (options *ApplyOptions) SetCreateOnly(createOnly bool) *ApplyOptions {
	options.createOnly = &createOnly
	return options
}

func (options *ApplyOptions) OwnerSetter() OwnerSetter {
	return options.setOwner
}

func (options *ApplyOptions) SetOwnerSetter(setOwner OwnerSetter) *ApplyOptions {
	options.setOwner = setOwner
	return options
}
