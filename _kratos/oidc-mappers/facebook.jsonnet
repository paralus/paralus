// scopes: email
// Issuer Url: https://www.facebook.com

local claims = std.extVar('claims');
{
  identity: {
    traits: {
      // The email might be empty if the user hasn't granted permissions for the email scope.
      [if "email" in claims then "email" else null]: claims.email,
      [if "given_name" in claims then "first_name" else null]: claims.given_name,
      [if "family_name" in claims then "last_name" else null]: claims.family_name,
    },
  },
}
