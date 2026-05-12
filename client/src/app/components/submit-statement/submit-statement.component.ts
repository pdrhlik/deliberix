import { Component, inject, input, output, signal } from "@angular/core";
import { FormsModule } from "@angular/forms";
import { IonButton, IonIcon, IonInput, IonNote, ToastController } from "@ionic/angular/standalone";
import { TranslatePipe, TranslateService } from "@ngx-translate/core";
import { addIcons } from "ionicons";
import { sendOutline } from "ionicons/icons";
import { StatementService } from "../../services/statement.service";

@Component({
  selector: "app-submit-statement",
  standalone: true,
  imports: [FormsModule, TranslatePipe, IonInput, IonButton, IonIcon, IonNote],
  templateUrl: "./submit-statement.component.html",
  styleUrls: ["./submit-statement.component.scss"],
})
export class SubmitStatementComponent {
  private statementService = inject(StatementService);
  private toastController = inject(ToastController);
  private translate = inject(TranslateService);

  surveySlug = input.required<string>();
  charMin = input<number>(20);
  charMax = input<number>(150);
  moderationEnabled = input<boolean>(true);

  statementSubmitted = output<void>();

  newText = "";
  submitted = signal(false);

  constructor() {
    addIcons({ sendOutline });
  }

  get charCount(): number {
    return this.newText.length;
  }

  get isValid(): boolean {
    const len = this.newText.length;
    return len >= this.charMin() && len <= this.charMax();
  }

  async submit() {
    if (!this.isValid) return;

    await this.statementService.submitStatement(this.surveySlug(), this.newText.trim());
    this.newText = "";
    this.submitted.set(true);
    this.statementSubmitted.emit();

    const messageKey = this.moderationEnabled()
      ? "statement.submitted-for-moderation"
      : "statement.submitted-live";
    const toast = await this.toastController.create({
      message: this.translate.instant(messageKey),
      duration: 3000,
      color: "success",
      position: "bottom",
    });
    await toast.present();

    setTimeout(() => this.submitted.set(false), 3000);
  }
}
