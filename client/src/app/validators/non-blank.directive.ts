import { Directive } from "@angular/core";
import { AbstractControl, NG_VALIDATORS, ValidationErrors, Validator } from "@angular/forms";

@Directive({
  selector: "[nonBlank]",
  standalone: true,
  providers: [{ provide: NG_VALIDATORS, useExisting: NonBlankDirective, multi: true }],
})
export class NonBlankDirective implements Validator {
  validate(control: AbstractControl): ValidationErrors | null {
    if (control.value == null || control.value === "") return null;
    return String(control.value).trim() === "" ? { required: true } : null;
  }
}
