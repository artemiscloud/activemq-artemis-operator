package selectors

import (
	"k8s.io/apimachinery/pkg/labels"
)

const (
	LabelAppKey = "application"

	LabelResourceKey = "AMQBroker"
)

// Set labels in a map
func LabelsForAMQBroker(name string) map[string]string {
	return map[string]string{
		LabelAppKey:      name + "-app",
		LabelResourceKey: name,
	}
}

// return a selector that matches resources for a AMQBroker resource
func ResourcesByAMQBrokerName(name string) labels.Selector {
	set := map[string]string{
		LabelAppKey:      name,
		//LabelResourceKey: name,
	}
	return labels.SelectorFromSet(set)
}

