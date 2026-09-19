package routes

import (
	"byteport/models"
	"log"
)

func addNewProject(project models.Project) error {
	log.Printf("project %s: writing row", project.UUID)
	return models.DB.Create(&project).Error
}

func removeProject(project models.Project) error {
	log.Printf("project %s: deleting row", project.UUID)
	return models.DB.Delete(&project).Error
}
