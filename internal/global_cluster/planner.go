package global_cluster

import "fmt"

type SecondaryClusterChangePlan struct {
	Add    []GlobalClusterMemberSpec
	Delete []GlobalClusterMember
}

func (c *GlobalCluster) PlanSecondaryClusterChange(desired []GlobalClusterMemberSpec) (SecondaryClusterChangePlan, error) {
	if len(desired) == 0 {
		return SecondaryClusterChangePlan{}, fmt.Errorf("cannot express empty global cluster members with the current Global Cluster API")
	}

	primary, secondaries := c.primaryAndSecondaries()
	if primary == nil {
		return SecondaryClusterChangePlan{}, fmt.Errorf("cannot express update for global cluster %s because the API response has no primary member", c.GlobalClusterID)
	}

	if desired[0].ClusterName != primary.ClusterName || desired[0].RegionID != primary.RegionID {
		return SecondaryClusterChangePlan{}, fmt.Errorf("cannot express primary member changes with the current Global Cluster API; replace the resource instead")
	}

	desiredSecondaries := desired[1:]
	toDelete := secondariesToDelete(secondaries, desiredSecondaries)
	toAdd := secondariesToAdd(secondaries, desiredSecondaries)
	if len(toDelete) > 0 && len(toAdd) > 0 {
		return SecondaryClusterChangePlan{}, fmt.Errorf("cannot express secondary replacement in one update with the current Global Cluster API; remove old secondary members first, apply, then add new members")
	}

	return SecondaryClusterChangePlan{Add: toAdd, Delete: toDelete}, nil
}

func (c *GlobalCluster) primaryAndSecondaries() (*GlobalClusterMember, []GlobalClusterMember) {
	var primary *GlobalClusterMember
	secondaries := make([]GlobalClusterMember, 0)
	for i := range c.Clusters {
		switch c.Clusters[i].Role {
		case GlobalClusterMemberRolePrimary:
			primary = &c.Clusters[i]
		case GlobalClusterMemberRoleSecondary:
			secondaries = append(secondaries, c.Clusters[i])
		}
	}
	return primary, secondaries
}

func secondariesToDelete(current []GlobalClusterMember, desired []GlobalClusterMemberSpec) []GlobalClusterMember {
	desiredCounts := specCounts(desired)
	result := make([]GlobalClusterMember, 0)
	for _, member := range current {
		key := memberKey(member.ClusterName, member.RegionID)
		if desiredCounts[key] > 0 {
			desiredCounts[key]--
			continue
		}
		result = append(result, member)
	}
	return result
}

func secondariesToAdd(current []GlobalClusterMember, desired []GlobalClusterMemberSpec) []GlobalClusterMemberSpec {
	currentCounts := memberCounts(current)
	result := make([]GlobalClusterMemberSpec, 0)
	for _, spec := range desired {
		key := memberKey(spec.ClusterName, spec.RegionID)
		if currentCounts[key] > 0 {
			currentCounts[key]--
			continue
		}
		result = append(result, spec)
	}
	return result
}

func specCounts(specs []GlobalClusterMemberSpec) map[string]int {
	counts := make(map[string]int, len(specs))
	for _, spec := range specs {
		counts[memberKey(spec.ClusterName, spec.RegionID)]++
	}
	return counts
}

func memberCounts(members []GlobalClusterMember) map[string]int {
	counts := make(map[string]int, len(members))
	for _, member := range members {
		counts[memberKey(member.ClusterName, member.RegionID)]++
	}
	return counts
}

func memberKey(clusterName string, regionID string) string {
	return clusterName + "\x00" + regionID
}
