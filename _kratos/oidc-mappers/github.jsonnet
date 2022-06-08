// scopes: user:email
// Issuer Url: https://github.com/login/oauth/authorize

local claims = {
  email_verified: false
} + std.extVar('claims');

local fName = if "name" in claims && claims.name!=null && std.length(std.findSubstr(" ", claims.name)) > 0 then std.splitLimit(claims.name, " ", 1)[0] else "Paralus";
local lName = if "name" in claims && claims.name!=null && std.length(std.findSubstr(" ", claims.name)) > 0 then std.splitLimit(claims.name, " ", 1)[1] else "User";

{
  identity: {
    traits: {
      // Allowing unverified email addresses enables account
      // enumeration attacks, especially if the value is used for
      // e.g. verification or as a password login identifier.
      //
      // Therefore we only return the email if it (a) exists and (b) is marked verified
      // by GitHub.
      [if "email" in claims && claims.email_verified then "email" else null]: claims.email,
      first_name: fName,
      last_name: lName,
    },
  },
}
