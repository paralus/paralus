window.onload = function() {
  //<editor-fold desc="Changeable Configuration Block">

  // the following lines will be replaced by docker/configurator, when it runs in a docker-container
  window.ui = SwaggerUIBundle({
    urls: [   
        { url: "https://github.com/paralus/paralus/blob/gh-pages/apis/partner.swagger.json", name: "Partner" },  
        { url: "https://github.com/paralus/paralus/blob/gh-pages/apis/organization.swagger.json", name: "Organization" }  
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