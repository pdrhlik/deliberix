import {
  Component,
  HostListener,
  Input,
  ViewEncapsulation,
  inject,
} from "@angular/core";
import { QRCodeComponent } from "angularx-qrcode";
import { FullscreenQrService } from "../../services/fullscreen-qr.service";

@Component({
  selector: "app-fullscreen-qr",
  standalone: true,
  imports: [QRCodeComponent],
  template: `
    <div class="fullscreen-qr-wrap" (click)="dismiss()">
      <qrcode
        [qrdata]="url"
        [width]="1024"
        [margin]="2"
        [errorCorrectionLevel]="'M'"
        [elementType]="'svg'"
      ></qrcode>
    </div>
  `,
  styleUrls: ["./fullscreen-qr.component.scss"],
  encapsulation: ViewEncapsulation.None,
})
export class FullscreenQrComponent {
  private service = inject(FullscreenQrService);

  @Input({ required: true }) url = "";

  @HostListener("document:keydown.escape")
  dismiss() {
    this.service.hide();
  }
}
