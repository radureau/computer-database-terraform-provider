resource "computer-database_company" "oxyl" {
  id = "oxyl"
  name = "Oxyl"

  computer_models = [
    {
      id = "oxyl-00"
      name = "Capico 00"
      release = 2000
    },
    {
      id = "oxyl-01"
      name = "MNF 21"
      release = 2021
    }
  ]
}

