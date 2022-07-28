package kubeapi

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/util/kubernetes/apiserver"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	fake4 "k8s.io/client-go/kubernetes/typed/apps/v1/fake"
	fake2 "k8s.io/client-go/kubernetes/typed/batch/v1/fake"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	fakecorev1 "k8s.io/client-go/kubernetes/typed/core/v1/fake"
	fake3 "k8s.io/client-go/kubernetes/typed/extensions/v1beta1/fake"
	"k8s.io/client-go/rest"
	core "k8s.io/client-go/testing"
	k8stesting "k8s.io/client-go/testing"
	"regexp"
	"strings"
	"testing"
)

type clientSetHTTP struct {
	*fake.Clientset
}

func (sc clientSetHTTP) CoreV1() corev1.CoreV1Interface {
	return coveV1NoRestClient{sc.Clientset.CoreV1().(*fakecorev1.FakeCoreV1)}
}

type coveV1NoRestClient struct {
	*fakecorev1.FakeCoreV1
}

func (c coveV1NoRestClient) RESTClient() rest.Interface {
	return nil
}

func getReactor(fakeClient *fake.Clientset, group string) *k8stesting.Fake {
	switch group {
	case "":
		return fakeClient.CoreV1().(*fakecorev1.FakeCoreV1).Fake
	case "batch":
		return fakeClient.BatchV1().(*fake2.FakeBatchV1).Fake
	case "extensions":
		return fakeClient.ExtensionsV1beta1().(*fake3.FakeExtensionsV1beta1).Fake
	case "apps":
		return fakeClient.AppsV1().(*fake4.FakeAppsV1).Fake
	default:
		return nil
	}
}

// MockAPIClient create a K8s API Client that can return errors for specified resource rules
func MockAPIClient(restrictRules []Rule) *apiserver.APIClient {
	fakeClient := fake.NewSimpleClientset()

	for _, rule := range restrictRules {
		reactor := getReactor(fakeClient, rule.Group)
		for _, verb := range rule.Verbs {
			reactor.
				PrependReactor(verb, rule.ResourceName, func(action core.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, fmt.Errorf("no permission to %s %s/%s", verb, rule.Group, rule.ResourceName)
				})
		}
	}

	x := clientSetHTTP{fakeClient}

	return &apiserver.APIClient{
		Cl: x,
	}
}

var ruleDescriptionRegexp = regexp.MustCompile(`^((?P<group>\w+)/)?(?P<name>\w+)\+(?P<verb>[\w,]+)`)

// Rule keeps verbs (get,list...) along with a resource (in group) they are applied to
type Rule struct {
	Group        string
	ResourceName string
	Verbs        []string
}

func parseRule(description string) Rule {
	match := ruleDescriptionRegexp.FindStringSubmatch(description)
	if len(match) == 0 {
		return Rule{"", "", nil}
	}

	resGroup := match[2]
	resName := match[3]
	verbs := strings.Split(match[4], ",")

	return Rule{resGroup, resName, verbs}
}

func TestRuleParsing(t *testing.T) {
	simpleRule := parseRule("resource+action")
	assert.Equal(t, Rule{"", "resource", []string{"action"}}, simpleRule)

	twoActions := parseRule("resource+action1,action2")
	assert.Equal(t, Rule{"", "resource", []string{"action1", "action2"}}, twoActions)

	withGroup := parseRule("group/resource+action1")
	assert.Equal(t, Rule{"group", "resource", []string{"action1"}}, withGroup)

	groupAndActions := parseRule("group/resource+action1,action2")
	assert.Equal(t, Rule{"group", "resource", []string{"action1", "action2"}}, groupAndActions)
}

func parseRules(ruleDescs []string) []Rule {
	var rules []Rule
	for _, rule := range ruleDescs {
		rules = append(rules, parseRule(rule))
	}
	return rules
}
