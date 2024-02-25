package main

type Import struct {
	IsDefaultImport bool
	Module          string
	Path            string
	Extension       string
	FileName        string
}

func NewImport(path string) Import {
	return Import{
		Module:          "",
		Path:            path,
		Extension:       "",
		IsDefaultImport: false,
	}
}

func (i *Import) SetImportModule(module string) {
	i.Module = module
}

func (i *Import) SetExtension(extension string) {
	i.Extension = extension
}

func (i *Import) SetFileName(fileName string) {
	i.FileName = fileName
}
