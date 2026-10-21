import { postAction } from '$lib/api';
import type { Site } from './sites';

export function addDomain(site: Site, name: string) {
  return postAction(
    `/api/sites/${encodeURIComponent(site.domain)}/domain:add?name=${encodeURIComponent(name)}`
  );
}

export function editDomain(site: Site, oldName: string, newName: string) {
  return postAction(
    `/api/sites/${encodeURIComponent(site.domain)}/domain:edit?old=${encodeURIComponent(
      oldName
    )}&new=${encodeURIComponent(newName)}`
  );
}

export function removeDomain(site: Site, name: string) {
  return postAction(
    `/api/sites/${encodeURIComponent(site.domain)}/domain:remove?name=${encodeURIComponent(name)}`
  );
}
