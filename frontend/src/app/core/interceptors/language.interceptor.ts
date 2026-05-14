import { HttpInterceptorFn } from '@angular/common/http';
import { inject } from '@angular/core';
import { LanguageDetector } from '../services/language-detector.service';

export const languageInterceptor: HttpInterceptorFn = (req, next) => {
  const languageDetector = inject(LanguageDetector);
  
  // Agregar header X-Language con el idioma actual
  const lang = languageDetector.getLanguage();
  const modifiedReq = req.clone({
    setHeaders: {
      'X-Language': lang
    }
  });
  
  return next(modifiedReq);
};
