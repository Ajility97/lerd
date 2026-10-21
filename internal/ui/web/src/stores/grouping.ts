import { postAction } from '$lib/api';
import type { Site } from './sites';

/** Group an existing secondary site under main at the given subdomain label. */
export function assignGroup(main: Site, secondaryDomain: string, label: string, shareDB = false) {
  return postAction(
    `/api/sites/${encodeURIComponent(main.domain)}/group:assign?secondary=${encodeURIComponent(
      secondaryDomain
    )}&label=${encodeURIComponent(label)}${shareDB ? '&share_db=1' : ''}`
  );
}

/** Toggle whether a secondary shares the group main's database. */
export function setGroupSharedDB(secondary: Site, share: boolean) {
  return postAction(
    `/api/sites/${encodeURIComponent(secondary.domain)}/group:set-db?share=${share ? '1' : '0'}`
  );
}

/** Ungroup a secondary site, restoring its standalone domain. */
export function unassignGroup(secondary: Site) {
  return postAction(`/api/sites/${encodeURIComponent(secondary.domain)}/group:unassign`);
}

/** Change the subdomain label of a secondary site. */
export function setGroupLabel(secondary: Site, label: string) {
  return postAction(
    `/api/sites/${encodeURIComponent(secondary.domain)}/group:set-label?label=${encodeURIComponent(
      label
    )}`
  );
}

/** Dissolve the whole group the given site belongs to. */
export function dissolveGroup(site: Site) {
  return postAction(`/api/sites/${encodeURIComponent(site.domain)}/group:remove`);
}
