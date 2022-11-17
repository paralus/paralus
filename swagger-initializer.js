window.onload = function() {
  //<editor-fold desc="Changeable Configuration Block">

  // the following lines will be replaced by docker/configurator, when it runs in a docker-container
  window.ui = SwaggerUIBundle({
    urls: [   
        { url: "https://paralus.github.io/paralus/apis/partner.swagger.json", name: "Partner" },  
        { url: "https://paralus.github.io/paralus/apis/organization.swagger.json", name: "Organization" },
        { url: "https://paralus.github.io/paralus/apis/project.swagger.json", name: "Project" },
        { url: "https://paralus.github.io/paralus/apis/cluster.swagger.json", name: "Cluster" },
        { url: "https://paralus.github.io/paralus/apis/user.swagger.json", name: "User" },
        { url: "https://paralus.github.io/paralus/apis/group.swagger.json", name: "Group" },
        { url: "https://paralus.github.io/paralus/apis/role.swagger.json", name: "Role" },
        { url: "https://paralus.github.io/paralus/apis/rolepermission.swagger.json", name: "RolePermission" },
        { url: "https://paralus.github.io/paralus/apis/kubectl_cluster.swagger.json", name: "KubectlSettings" },
        { url: "https://paralus.github.io/paralus/apis/oidc_provider.swagger.json", name: "OIDCProvider" },
        { url: "https://paralus.github.io/paralus/apis/auditlog.swagger.json", name: "AuditLog" },
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
