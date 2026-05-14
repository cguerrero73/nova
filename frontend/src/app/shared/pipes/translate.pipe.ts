import { Pipe, PipeTransform, inject } from '@angular/core';
import { TranslationService } from '../services/translation.service';

@Pipe({
  name: 'translate',
  standalone: true,
  pure: false, // Para que se actualice cuando cambian las traducciones
})
export class TranslatePipe implements PipeTransform {
  private readonly translationService = inject(TranslationService);
  
  private currentScreenId: string = '';
  private currentKey: string = '';

  transform(key: string, screenId: string): string {
    // Si cambió el screenId, cargar las traducciones
    if (screenId !== this.currentScreenId) {
      this.currentScreenId = screenId;
      this.translationService.load(screenId).subscribe();
    }
    
    this.currentKey = key;
    return this.translationService.t(screenId, key);
  }
}
