import { m } from '../paraglide/messages.js';

// startStepLabel maps a stage id from the start stream to its message.
export function startStepLabel(step: string): string {
  switch (step) {
    case 'machine':
      return m.dashboard_hero_startStep_machine();
    case 'preparing':
      return m.dashboard_hero_startStep_preparing();
    case 'images':
      return m.dashboard_hero_startStep_images();
    case 'units':
      return m.dashboard_hero_startStep_units();
    case 'dns':
      return m.dashboard_hero_startStep_dns();
    default:
      return '';
  }
}
