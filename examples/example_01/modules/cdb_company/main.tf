resource "computer-database_company" "company" {
  id = lower(var.name)
  name = var.name

  computer_models = [
    for idx,vm in var.computer_models:
    {
      id = format("%s-%d", lower(var.name), idx)
      name = vm.name
      release = vm.release != null ? vm.release : 2023
    }
  ]
}