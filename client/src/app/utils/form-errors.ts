import { NgForm } from "@angular/forms";

export function firstFormErrorKey(f: NgForm): string | null {
  for (const name of Object.keys(f.controls)) {
    const c = f.controls[name];
    if (!c.errors) continue;
    if (c.errors["email"]) return "auth.invalid-email";
    if (c.errors["matches"]) return "auth.passwords-no-match";
    if (c.errors["minlength"]) return "auth.password-too-short";
    if (c.errors["required"]) return "auth.required-fields";
  }
  return null;
}
