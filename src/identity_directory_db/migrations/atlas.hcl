data "external_schema" "gorm-postgres" {
  program = [
    "go",
    "run",
    "./loader",
    "-dialect",
    "postgres",
  ]
}

env "gorm-postgres" {
  src = data.external_schema.gorm-postgres.url
  dev = "docker://postgres/18/dev"
  migration {
    dir = "file://postgres"
    format = goose
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}
