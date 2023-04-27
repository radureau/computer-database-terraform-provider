module "oxyl_mirror" {
  source = "./modules/cdb_company"
  
  name = strrev(computer-database_company.oxyl.name)

  computer_models = [
    for cm in  computer-database_company.oxyl.computer_models:
    {
        name = strrev(cm.name)
        release = 2023
    }
  ]
}

output "mirror" {
    value = module.oxyl_mirror
}