package converter

import (
	serviceModel "PR/internal/model"
	repoModel "PR/internal/repository/team/model"
)

func FromRepo(members []*repoModel.TeamMember) []*serviceModel.TeamMember {
	serviceMembers := make([]*serviceModel.TeamMember, 0, len(members))

	for _, m := range members {
		serviceMembers = append(serviceMembers, &serviceModel.TeamMember{
			ID:       m.ID,
			Username: m.Username,
			IsActive: m.IsActive,
		})
	}
	return serviceMembers
}
