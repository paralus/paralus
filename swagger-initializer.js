window.onload = function() {
  //<editor-fold desc="Changeable Configuration Block">

  // the following lines will be replaced by docker/configurator, when it runs in a docker-container
  window.ui = SwaggerUIBundle({
    urls: [   
        { url: "https://paralus.github.io/paralus/apis/partner.swagger.json", name: "Partner" },  
        { url: "https://paralus.github.io/paralus/apis/Organization.swagger.json", name: "Organization" },
        { url: "https://paralus.github.io/paralus/apis/Project.swagger.json", name: "Project" },
    ],  
    "urls.primaryName": "Partner",
    dom_id: '#swagger-ui',
    deepLinking: true,
    presets: [
      SwaggerUIBundle.presets.apis,
      SwaggerUIStandalonePreset
    ],
    plugins: [
      SwaggerUIBundle.plugins.DownloadUrl
    ],
    layout: "StandaloneLayout"
  });

  //</editor-fold>
};
