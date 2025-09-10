package cache

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func FuzzSetAdd(f *testing.F) {
	f.Add("test-set", "test-member", "dep1")
	f.Add("", "", "")
	f.Add("unicode-set-🏠", "unicode-member-🔑", "unicode-dep-🎯")
	f.Add("set with spaces", "member with spaces", "dep with spaces")
	f.Add("set\nwith\nnewlines", "member\nwith\nnewlines", "dep\nwith\nnewlines")
	f.Add("set\x00\x01\x02", "member\x00\x01\x02", "dep\x00\x01\x02")
	f.Add(strings.Repeat("long-set-", 100), strings.Repeat("long-member-", 100), "dep")

	f.Fuzz(func(t *testing.T, setName, member, dependency string) {
		client, conn := loadMockRedis()
		defer client.Close()

		conn.Command(AddToSetCommand, setName, member).Expect(int64(1))
		conn.Command(AddToSetCommand, DependencyPrefix+dependency, setName).Expect(int64(1))
		conn.Command(MultiCommand).Expect("QUEUED")
		conn.Command(ExecuteCommand).Expect([]interface{}{int64(1)})

		ctx := context.Background()

		assert.NotPanics(t, func() {
			err := SetAdd(ctx, client, setName, member, dependency)
			assert.NoError(t, err)
		})
	})
}

func FuzzSetAddMany(f *testing.F) {
	f.Add("test-set", "member1,member2,member3")
	f.Add("", "")
	f.Add("unicode-set-🏠", "unicode-member1-🔑,unicode-member2-🚀")
	f.Add("set with spaces", "member with spaces")
	f.Add("set\nwith\nnewlines", "member\nwith\nnewlines")
	f.Add("binary-set", "member\x00\x01\x02")
	f.Add("long-set", strings.Repeat("long-member-", 50))

	f.Fuzz(func(t *testing.T, setName, membersStr string) {
		if membersStr == "" {
			return
		}

		members := strings.Split(membersStr, ",")

		client, conn := loadMockRedis()
		defer client.Close()

		args := make([]interface{}, len(members)+1)
		args[0] = setName
		for i, member := range members {
			args[i+1] = member
		}

		conn.Command(AddToSetCommand, args...).Expect(int64(len(members)))

		ctx := context.Background()

		assert.NotPanics(t, func() {
			err := SetAddMany(ctx, client, setName, stringSliceToInterfaceSlice(members)...)
			assert.NoError(t, err)
		})
	})
}

func FuzzSetIsMember(f *testing.F) {
	f.Add("test-set", "test-member")
	f.Add("", "")
	f.Add("unicode-set-🏠", "unicode-member-🔑")
	f.Add("set with spaces", "member with spaces")
	f.Add("set\nwith\nnewlines", "member\nwith\nnewlines")
	f.Add("set\x00\x01\x02", "member\x00\x01\x02")
	f.Add(strings.Repeat("long-set-", 50), strings.Repeat("long-member-", 50))

	f.Fuzz(func(t *testing.T, setName, member string) {
		client, conn := loadMockRedis()
		defer client.Close()

		conn.Command(IsMemberCommand, setName, member).Expect(int64(1))

		ctx := context.Background()

		assert.NotPanics(t, func() {
			result, err := SetIsMember(ctx, client, setName, member)
			if err == nil {
				assert.IsType(t, bool(true), result)
			}
		})
	})
}

func FuzzSetRemoveMember(f *testing.F) {
	f.Add("test-set", "test-member")
	f.Add("", "")
	f.Add("unicode-set-🏠", "unicode-member-🔑")
	f.Add("set with spaces", "member with spaces")
	f.Add("set\nwith\nnewlines", "member\nwith\nnewlines")
	f.Add("set\x00\x01\x02", "member\x00\x01\x02")
	f.Add(strings.Repeat("long-set-", 50), strings.Repeat("long-member-", 50))

	f.Fuzz(func(t *testing.T, setName, member string) {
		client, conn := loadMockRedis()
		defer client.Close()

		conn.Command(RemoveMemberCommand, setName, member).Expect(int64(1))

		ctx := context.Background()

		assert.NotPanics(t, func() {
			err := SetRemoveMember(ctx, client, setName, member)
			assert.NoError(t, err)
		})
	})
}

func FuzzSetMembers(f *testing.F) {
	f.Add("test-set", "member1,member2,member3")
	f.Add("", "")
	f.Add("unicode-set-🏠", "unicode-member1-🔑,unicode-member2-🚀")
	f.Add("set with spaces", "member with spaces")
	f.Add("empty-set", "")

	f.Fuzz(func(t *testing.T, setName, expectedMembersStr string) {
		var expectedMembers []string
		if expectedMembersStr != "" {
			expectedMembers = strings.Split(expectedMembersStr, ",")
		}
		client, conn := loadMockRedis()
		defer client.Close()

		expectedValues := make([]interface{}, len(expectedMembers))
		for i, member := range expectedMembers {
			expectedValues[i] = []byte(member)
		}

		conn.Command(MembersCommand, setName).Expect(expectedValues)

		ctx := context.Background()

		assert.NotPanics(t, func() {
			result, err := SetMembers(ctx, client, setName)
			if err == nil {
				assert.IsType(t, []string{}, result)
				if len(expectedMembers) > 0 {
					assert.Len(t, result, len(expectedMembers))
				}
			}
		})
	})
}

func FuzzSetOperationsRoundTrip(f *testing.F) {
	f.Add("test-set", "member1", "member2", "member3")
	f.Add("unicode-set-🏠", "unicode-member1-🔑", "unicode-member2-🚀", "unicode-member3-⭐")
	f.Add("set with spaces", "member1 with spaces", "member2 with spaces", "member3 with spaces")

	f.Fuzz(func(t *testing.T, setName, member1, member2, member3 string) {
		if member1 == "" && member2 == "" && member3 == "" {
			return
		}

		client, conn := loadMockRedis()
		defer client.Close()

		var members []string
		var interfaceMembers []interface{}
		var expectedValues []interface{}

		if member1 != "" {
			members = append(members, member1)
			interfaceMembers = append(interfaceMembers, member1)
			expectedValues = append(expectedValues, []byte(member1))
		}
		if member2 != "" && member2 != member1 {
			members = append(members, member2)
			interfaceMembers = append(interfaceMembers, member2)
			expectedValues = append(expectedValues, []byte(member2))
		}
		if member3 != "" && member3 != member1 && member3 != member2 {
			members = append(members, member3)
			interfaceMembers = append(interfaceMembers, member3)
			expectedValues = append(expectedValues, []byte(member3))
		}

		if len(members) == 0 {
			return
		}

		addArgs := make([]interface{}, len(members)+1)
		addArgs[0] = setName
		for i, member := range members {
			addArgs[i+1] = member
		}

		conn.Command(AddToSetCommand, addArgs...).Expect(int64(len(members)))

		for _, member := range members {
			conn.Command(IsMemberCommand, setName, member).Expect(int64(1))
		}

		conn.Command(MembersCommand, setName).Expect(expectedValues)

		ctx := context.Background()

		assert.NotPanics(t, func() {
			err := SetAddMany(ctx, client, setName, interfaceMembers...)
			assert.NoError(t, err)

			for _, member := range members {
				isMember, memberErr := SetIsMember(ctx, client, setName, member)
				if memberErr == nil {
					assert.True(t, isMember)
				}
			}

			allMembers, err := SetMembers(ctx, client, setName)
			if err == nil {
				assert.Len(t, allMembers, len(members))
			}
		})
	})
}

func FuzzSetAddWithDependencies(f *testing.F) {
	f.Add("test-set", "test-member", "dep1", "dep2")
	f.Add("unicode-set-🏠", "unicode-member-🔑", "unicode-dep1-🎯", "unicode-dep2-🚀")

	f.Fuzz(func(t *testing.T, setName, member, dep1, dep2 string) {
		client, conn := loadMockRedis()
		defer client.Close()

		conn.Command(AddToSetCommand, setName, member).Expect(int64(1))
		conn.Command(AddToSetCommand, DependencyPrefix+dep1, setName).Expect(int64(1))
		if dep2 != "" && dep2 != dep1 {
			conn.Command(AddToSetCommand, DependencyPrefix+dep2, setName).Expect(int64(1))
		}
		conn.Command(MultiCommand).Expect("QUEUED")
		conn.Command(ExecuteCommand).Expect([]interface{}{int64(1)})

		ctx := context.Background()
		dependencies := []string{dep1}
		if dep2 != "" && dep2 != dep1 {
			dependencies = append(dependencies, dep2)
		}

		assert.NotPanics(t, func() {
			err := SetAdd(ctx, client, setName, member, dependencies...)
			assert.NoError(t, err)
		})
	})
}

func stringSliceToInterfaceSlice(strings []string) []interface{} {
	interfaces := make([]interface{}, len(strings))
	for i, s := range strings {
		interfaces[i] = s
	}
	return interfaces
}
