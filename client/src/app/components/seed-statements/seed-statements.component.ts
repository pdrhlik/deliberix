import { Component, inject, input, OnInit, signal } from "@angular/core";
import { FormsModule } from "@angular/forms";
import { IonButton, IonIcon, IonInput, IonNote } from "@ionic/angular/standalone";
import { TranslatePipe } from "@ngx-translate/core";
import { addIcons } from "ionicons";
import { addOutline } from "ionicons/icons";
import { Statement } from "../../models/statement.model";
import { StatementService } from "../../services/statement.service";

@Component({
  selector: "app-seed-statements",
  standalone: true,
  imports: [FormsModule, TranslatePipe, IonInput, IonButton, IonIcon, IonNote],
  templateUrl: "./seed-statements.component.html",
  styleUrls: ["./seed-statements.component.scss"],
})
export class SeedStatementsComponent implements OnInit {
  private statementService = inject(StatementService);

  surveySlug = input.required<string>();
  charMin = input<number>(20);
  charMax = input<number>(150);

  statements = signal<Statement[]>([]);
  newText = "";

  constructor() {
    addIcons({ addOutline });
  }

  ngOnInit() {
    this.loadStatements();
  }

  async loadStatements() {
    const items = await this.statementService.listStatements(this.surveySlug());
    this.statements.set(items);
  }

  get charCount(): number {
    return this.newText.length;
  }

  get isValid(): boolean {
    const len = this.newText.length;
    return len >= this.charMin() && len <= this.charMax();
  }

  async addStatement() {
    if (!this.isValid) return;

    await this.statementService.addSeedStatement(this.surveySlug(), this.newText.trim());
    this.newText = "";
    await this.loadStatements();
  }
}
