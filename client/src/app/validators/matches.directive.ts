import { Directive, Input, OnChanges, SimpleChanges } from "@angular/core";
import { AbstractControl, NG_VALIDATORS, ValidationErrors, Validator } from "@angular/forms";

@Directive({
  selector: "[matches]",
  standalone: true,
  providers: [{ provide: NG_VALIDATORS, useExisting: MatchesDirective, multi: true }],
})
export class MatchesDirective implements Validator, OnChanges {
  @Input("matches") matchValue: unknown = null;
  private onChange?: () => void;

  ngOnChanges(changes: SimpleChanges) {
    if (changes["matchValue"] && this.onChange) {
      this.onChange();
    }
  }

  registerOnValidatorChange(fn: () => void) {
    this.onChange = fn;
  }

  validate(control: AbstractControl): ValidationErrors | null {
    if (control.value == null || control.value === "") return null;
    return control.value === this.matchValue ? null : { matches: true };
  }
}
