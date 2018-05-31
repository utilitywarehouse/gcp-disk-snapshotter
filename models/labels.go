package models

type Label struct {
	Name  string
	Value string
}

type LabelList struct {
	Items []Label
}

type LabelListInterface interface {
	AddLabel(name, value string)
}

func (ll *LabelList) AddLabel(name, value string) {
	label := &Label{
		Name:  name,
		Value: value,
	}
	ll.Items = append(ll.Items, *label)
}
