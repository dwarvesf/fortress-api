package request

type UploadContentRequest struct {
	Module      UploadModule `form:"module" binding:"required"`
	ContentType ContentType  `form:"contentType" binding:"required"`
	EmployeeID  string       `form:"employeeID"`
	ProjectID   string       `form:"projectID"`
}

type UploadModule string

const (
	UploadModuleResources UploadModule = "resources"
	UploadModuleProjects  UploadModule = "projects"
	UploadModuleEmployees UploadModule = "employees"
)

type ContentType string

const (
	ContentTypeImages    ContentType = "images"
	ContentTypeDocuments ContentType = "documents"
)
