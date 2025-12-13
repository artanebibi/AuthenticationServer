package service

import (
	"AuthServer/internal/domain/models"
	"AuthServer/internal/repository"
)

type IProjectService interface {
	FindAll() ([]models.Project, error)
	FindById(id string) (*models.Project, error)
	FindByName(name string) (*models.Project, error)
	Save(project models.Project) (*models.Project, error)
	Update(project models.Project) error
	Delete(id string) error
}

type ProjectService struct {
	projectRepository repository.IProjectRepository
}

func NewProjectService(repo repository.IProjectRepository) *ProjectService {
	return &ProjectService{
		projectRepository: repo,
	}
}

func (p *ProjectService) FindAll() ([]models.Project, error) {
	return p.projectRepository.FindAll()
}

func (p *ProjectService) FindById(id string) (*models.Project, error) {
	project, err := p.projectRepository.FindById(id)
	if err != nil {
		return nil, err
	}
	return project, nil
}

func (p *ProjectService) FindByName(name string) (*models.Project, error) {
	project, err := p.projectRepository.FindByName(name)
	if err != nil {
		return nil, err
	}
	return project, nil
}

func (p *ProjectService) Save(project models.Project) (*models.Project, error) {
	err := p.projectRepository.Save(project)
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (p *ProjectService) Update(project models.Project) error {
	return p.projectRepository.Update(project)
}

func (p *ProjectService) Delete(id string) error {
	return p.projectRepository.Delete(id)
}
