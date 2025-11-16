package converter

import (
	serviceModel "PR/internal/model"
	repoModel "PR/internal/repository/user/model"
)

func FromRepo(u *repoModel.User) *serviceModel.User {
	return &serviceModel.User{
		ID:       u.ID,
		TeamName: u.TeamName,
		Username: u.Username,
		IsActive: u.IsActive,
	}
}

func FromRepoList(users []*repoModel.User) []*serviceModel.User {
	serviceUser := make([]*serviceModel.User, 0, len(users))

	for _, u := range users {
		serviceUser = append(serviceUser, FromRepo(u))
	}
	return serviceUser
}
