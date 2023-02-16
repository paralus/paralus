// Issuer Url: <Okta domain url>
// scopes: email, profile, openid

local claims = std.extVar('claims');

local fName = if "name" in claims && claims.name!=null && std.length(std.findSubstr(" ", claims.name)) > 0 then std.splitLimit(claims.name, " ", 1)[0] else "Paralus";
local lName = if "name" in claims && claims.name!=null && std.length(std.findSubstr(" ", claims.name)) > 0 then std.splitLimit(claims.name, " ", 1)[1] else "User";

{
  identity: {
    traits: {
      email: claims.email,
      first_name: fName,
      last_name: lName,
      [if "groups" in claims.raw_claims then "idp_groups" else null]: claims.raw_claims.groups,
    },
  },
}
