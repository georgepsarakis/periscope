data "external_schema" "gorm" {
  program = [
    "go",
    "run",
    "-mod=mod",
    "ariga.io/atlas-provider-gorm",
    "load",
    "--path", "./repository/rdbms",
    "--dialect", "postgres",
  ]
}

variable "url" {
  type = string
  default = getenv("DATABASE_URL")
}

env "gorm" {
  src = data.external_schema.gorm.url
  dev = "docker://postgres/16/periscope"
  url = var.url
  migration {
    dir = "file://migrations"
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}