import { bootstrapApplication } from '@angular/platform-browser';
import { provideRouter } from '@angular/router';
import { provideHttpClient, withInterceptors } from '@angular/common/http';
import { provideAnimations } from '@angular/platform-browser/animations';
import { AppComponent } from './app/app.component';
import { routes } from './app/app.routes';
import { environment } from './environments/environment';
import { languageInterceptor } from './app/core/interceptors/language.interceptor';
import { errorInterceptor } from './app/core/interceptors/error.interceptor';

async function bootstrap() {
  if (environment.useMock) {
    const { worker } = await import('./mocks/browser');
    await worker.start({ 
      onUnhandledRequest: 'bypass',
      serviceWorker: {
        url: '/mockServiceWorker.js',
      }
    });
  }

  bootstrapApplication(AppComponent, {
    providers: [
      provideRouter(routes),
      provideHttpClient(withInterceptors([languageInterceptor, errorInterceptor])),
      provideAnimations(),
    ],
  }).catch((err) => console.error(err));
}

bootstrap();
