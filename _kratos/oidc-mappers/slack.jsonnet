// scopes: identity.basic, identity.email
// Issuer Url: https://slack.com

local claims = {
  email_verified: true
} + std.extVar('claims');

local fName = if "name" in claims && claims.name!=null && std.length(std.findSubstr(" ", claims.name)) > 0 then std.splitLimit(claims.name, " ", 1)[0] else "Paralus";
local lName = if "name" in claims && claims.name!=null && std.length(std.findSubstr(" ", claims.name)) > 0 then std.splitLimit(claims.name, " ", 1)[1] else "User";

{
  identity: {
    traits: {
      // Allowing unverified email addresses enables account
      // enumeration attacks,  if the value is used for
      // verification or as a password login identifier.
      //
      // It's assumed that Slack requires an email to be verified to be accessible via OAuth (because they don't provide a email_verified field).
      email: claims.email,
      first_name: fName,
      last_name: lName,
    },
  },
}
