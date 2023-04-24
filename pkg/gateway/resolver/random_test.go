package resolver

import (
	"fmt"
	"net/url"
	"testing"

	"gotest.tools/v3/assert"
)

func Test_RandomResolver(t *testing.T) {

	funcName := "helloworld"

	resolver := NewResolver(RandomResolverFunc())
	for i := 0; i < 10; i++ {
		targetURl := fmt.Sprintf("http://asdfsad:80%d", i)
		u, _ := url.Parse(targetURl)
		assert.NilError(t, resolver.Add(funcName, u))
	}

	for i := 0; i < 10; i++ {
		u, ok := resolver.Get(funcName)
		if ok {
			fmt.Println(u)
		}
	}
	fmt.Println("===========", resolver.Len(funcName))
	for i := 0; i < 5; i++ {
		u, _ := resolver.Get(funcName)
		assert.NilError(t, resolver.Delete(funcName, u))
	}
	fmt.Println("===========", resolver.Len(funcName))
	for i := 0; i < 10; i++ {
		u, ok := resolver.Get(funcName)
		if ok {
			fmt.Println(u)
		}
	}

}
