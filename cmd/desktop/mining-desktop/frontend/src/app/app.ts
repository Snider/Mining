import { Component, signal, OnInit } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { GreetService } from '@bindings/changeme';
import { Events, Window } from '@wailsio/runtime';

@Component({
  selector: 'app-root',
  imports: [RouterOutlet],
  templateUrl: './app.html',
  styleUrl: './app.scss'
})
export class App implements OnInit {
  protected readonly title = signal('wails-angular-template');
  protected readonly result = signal('Please enter your name below 👇');
  protected readonly time = signal('Listening for Time event...');
  protected name = signal('');

  ngOnInit() {
    Events.On('time', (time: any) => {
      this.time.set(time.data);
    });
  }

  doGreet() {
    let name = this.name();
    if (!name) {
      name = 'anonymous';
    }
    GreetService.Greet(name).then((result: string) => {
      this.result.set(result);
    }).catch((err: any) => {
      console.log(err);
    });
  }

  openLink(url: string) {
    Events.Emit('openLink', url).then(r => console.log(r));
  }
}
