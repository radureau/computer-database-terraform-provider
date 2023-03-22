#!/usr/bin/env node
const http = require('http');
const fs = require('fs');

{
  const host = '127.0.0.1';
  const port = 8080;
  
  var config = {
    host,
    port,
  }
}

{
  const database = JSON.parse(fs.readFileSync(`${__dirname}/database.json`, 'utf8'));

  const computerModelFromDB = (dbComputerModel, dbCompany) => {
    let company = {...dbCompany}
    delete company.computerModels;
    return {...dbComputerModel, company}
  }
  const companyFromDB = (dbCompany) => {
    let company = {...dbCompany}
    company.computerModels = Object.values(dbCompany.computerModels).map(dbCM => computerModelFromDB(dbCM, dbCompany));
    return company;
  }

  const listCompanies = () => Object.values(database.companies).map(companyFromDB);
  
  const getDBCompany = (companyId) => companyId in database.companies ? database.companies[companyId] : undefined;
  const getCompany = (companyId) => companyId in database.companies ? companyFromDB(database.companies[companyId]) : undefined;
  
  const listCompanyComputerModels = (companyId) => {
    const dbCompany = getDBCompany(companyId);
    if (!dbCompany) return undefined;
    return Object.values(dbCompany.computerModels).map(computerModel => computerModelFromDB(computerModel, dbCompany));
  }
  
  const getCompanyComputerModel = (companyId, computerModelId) => {
    const dbCompany = getDBCompany(companyId);
    if (!dbCompany) return undefined;
    return computerModelId in dbCompany.computerModels ? 
      computerModelFromDB(
        database.companies[companyId].computerModels[computerModelId],
        dbCompany,
      )
      : undefined
  }

  const addCompany = (company) => {
    if (company.id in database.companies) return 'company.id not available';
    company.computerModels = company.computerModels.reduce( (db, cm) => { db[cm.id] = cm; return db }, {})
    database.companies[company.id] = company
  }
  const removeCompanyById = (companyId) => {
    if (!(companyId in database.companies)) return undefined
    const company = database.companies[companyId];
    delete database.companies[companyId];
    return company;
  }
  
  var store = {
    listCompanies,
    getCompany,
    listCompanyComputerModels,
    getCompanyComputerModel,

    addCompany,
    removeCompanyById,
  }
  store.updateCompany = (company) => {
    store.removeCompanyById(company.id)
    return store.addCompany(company)
  }
  store.addCompanyComputerModel = (companyId, computerModel) => {
    let company = store.getCompany(companyId)
    company.computerModels.push(computerModel)
    return store.updateCompany(company)
  }
}

{
  const apiPrefix = '/api/v1';
  const apiBaseURL = `http://${config.host}:${config.port}${apiPrefix}`;
  
  const companyURI = (companyId) => `${apiBaseURL}/companies/${companyId}`;
  const computerModelURI = (companyId, computerModelId) => `${apiBaseURL}/companies/${companyId}/computer-models/${computerModelId}`;
  
  var api = {
    prefix: apiPrefix,
    baseURL: apiBaseURL,
    
    companyURI,
    computerModelURI,
  }
}

/*
  Serializer
*/
{
  const stringify = JSON.stringify;
  JSON.stringify = (value) => stringify(value, null, '\t');
  Object.defineProperty(Array.prototype, 'json', {
    value: function () {
      if (this.length > 0 && this.every(e => e.hasOwnProperty('json')))
        return '[' + this.map(e => e.json()).join(',\n') + ']';
      return JSON.stringify(this);
    }
  });

  const JSONcompany = (company) => {
    company = {...company}
    company.uri = () => api.companyURI(company.id);
    company.json = () => JSON.stringify({ ...company, uri: company.uri(), computerModels: company.computerModels.map(cm => api.computerModelURI(company.id, cm.id)) });
    return company
  }
  const JSONComputerModel = (computerModel) => {
    computerModel = {...computerModel}
    computerModel.uri = () => api.computerModelURI(computerModel.company.id, computerModel.id);
    computerModel.json = () => JSON.stringify({ ...computerModel, uri: computerModel.uri(), company: api.companyURI(computerModel.company.id) });
    return computerModel;
  }

  const fromJSONComputerModel = ({id,name,release}) => {
    const jsonComputerModel = {id,name,release}
    if (Object.values(jsonComputerModel).some(v => !v)) return undefined;
    return jsonComputerModel;
  }
  const fromJSONCompany = ({id,name,location, computerModels}) => {
    const jsonCompany = {id,name,location}
    if (Object.values(jsonCompany).some(v => !v)) return undefined;
    jsonCompany.computerModels =  computerModels ?
      computerModels.map(fromJSONComputerModel).filter(Boolean)
      : []
    ;
    return jsonCompany;
  }

  var json = {
    company: JSONcompany,
    JSONComputerModel: JSONComputerModel,
  }
  var jsonParse = {
    company: fromJSONCompany,
    computerModel: fromJSONComputerModel,
  }
}

{
  const REST = {
    'companies': {
      'GET': (req, res) => {
        const companies = store.listCompanies();
        res.end(companies.map(json.company).json());
      },
      'POST': (req, res) => {
        const company = jsonParse.company(req.body);
        if (!company) { res.writeHead(400).end(); return }
        let err = store.addCompany(company);
        if (err) { res.writeHead(412, err).end(err); return }
        res.writeHead(201).end();
      },
    },
    'companies/:companyId': {
      'GET': (req, res, { companyId }) => {
        const company = store.getCompany(companyId)
        if (!company) { res.writeHead(404, `company ${companyId} not found`).end(); return }
        res.end(json.company(company).json());
      },
      'PUT': (req, res, { companyId }) => {
        const company = store.getCompany(companyId)
        if (!company) { res.writeHead(404, `company ${companyId} not found`).end(); return }
        const updatedCompany = jsonParse.company(req.body)
        if (!updatedCompany || updatedCompany.id !== company.id) { res.writeHead(400).end(); return }
        let err = store.updateCompany(updatedCompany);
        if (err) { res.writeHead(412, err).end(err); return }
        res.writeHead(200).end();
      },
      'DELETE': (req, res, { companyId }) => {
        const company = store.removeCompanyById(companyId);
        if (!company) { res.writeHead(404, `company ${companyId} not found`).end(); return }
        res.writeHead(204).end();
      },
    },
    'companies/:companyId/computer-models': {
      'GET': (req, res, { companyId }) => {
        const computerModels = store.listCompanyComputerModels(companyId)
        if (!computerModels) { res.writeHead(404, `computer models from company ${companyId} not found`).end(); return }
        res.end(computerModels.map(json.JSONComputerModel).json());
      },
      'POST': (req, res, { companyId }) => {
        const computerModel = jsonParse.computerModel(req.body);
        if (!computerModel) { res.writeHead(400).end(); return }
        
        let err = store.addCompanyComputerModel(companyId, computerModel);

        if (err) { res.writeHead(412, err).end(err); return }
        res.writeHead(201).end();
      },
    },
    'companies/:companyId/computer-models/:computerModelId': {
      'GET': (req, res, { companyId, computerModelId }) => {
        const computerModel = store.getCompanyComputerModel(companyId, computerModelId);
        if (!computerModel) { res.writeHead(404, `computer model ${computerModelId} not found in company ${companyId}`).end(); return }
        res.end(json.JSONComputerModel(computerModel).json());
      },
    },
  }
  var routes = REST;
}

/*
  HTTP Server and Middlewares
*/
{
  const _routesRegex = {}
  for (let route in routes) {
    const route_as_regex = route.replace(/:(?<param>[A-z_]+[A-z_0-9]*)/g, function () {
      const { param } = arguments[arguments.length - 1]; // replace callback function has variadic parameters and regex named captured groups is the last parameter
      return `(?<${param}>[^/]+)`
    })
    _routesRegex[route] = new RegExp(route_as_regex + '/?$');
  }
  
  const server = http.createServer((req, res) => {
  
    let path = req.url.replace(api.prefix, '');
  
    const routeRegex = Object.entries(_routesRegex).find(([_, re]) => re.test(path));
    if (!routeRegex) { res.writeHead(404, 'route not found').end(); return }
    
    const [route, re] = routeRegex
    req.params = re.exec(path).groups
    const handlers = routes[route]
    if (!(req.method in handlers)) { res.writeHead(404, 'route not found').end(); return }
    const handler = handlers[req.method]
    
    res.setHeader('Content-Type', 'application/json')
    if (['POST','PUT','PATCH'].includes(req.method)) {
      let data = '';
      req.on('data', chunk => {
        data += chunk;
      })
      req.on('end', () => {
        try {
          req.body = data ? JSON.parse(data) : {}
        } catch(e) {
          res.writeHead(400, e.message).end()
          return
        }
        handler(req, res, req.params);
      })
    } else {
      handler(req, res, req.params);
    }
  });

  var main = () => server.listen(config.port, config.host, () => {
    console.log(`Server running at ${api.baseURL}`);
  });
}

main()