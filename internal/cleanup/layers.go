package cleanup

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/geodro/lerd/internal/podman"
)

// historySizes is the seam tests override: the size of every history step of
// one image, newest first, the order podman history lists them in.
var historySizes = podmanHistorySizes

func podmanHistorySizes(id string) ([]int64, error) {
	out, err := podman.Run("history", "--format", "json", id)
	if err != nil {
		return nil, err
	}
	var rows []struct {
		Size int64 `json:"size"`
	}
	if err := json.Unmarshal([]byte(out), &rows); err != nil {
		return nil, fmt.Errorf("parsing history of %s: %w", id, err)
	}
	sizes := make([]int64, len(rows))
	for i, r := range rows {
		sizes[i] = r.Size
	}
	return sizes, nil
}

// layerSizesFrom pairs an image's layers, bottom first, with its history sizes,
// newest first. Only steps that wrote something are layers, so a real layer of
// zero bytes breaks the pairing, and then the count check refuses the lot.
func layerSizesFrom(layers []string, history []int64) (map[string]int64, bool) {
	var written []int64
	for i := len(history) - 1; i >= 0; i-- {
		if history[i] > 0 {
			written = append(written, history[i])
		}
	}
	if len(written) != len(layers) {
		return nil, false
	}
	sizes := make(map[string]int64, len(layers))
	for i, l := range layers {
		sizes[l] = written[i]
	}
	return sizes, true
}

// creditFreedLayers sets each image target to the bytes of layers no staying
// image holds, since podman's unique size reads a 1 GB FPM image as a few KB.
// Ancestors go first; a target whose sizes can't be read keeps its estimate.
func creditFreedLayers(targets []Target, imgs []image) {
	removing := targetImages(targets)
	if len(removing) == 0 {
		return
	}
	var ids []string
	for id := range removing {
		ids = append(ids, id)
	}
	for _, img := range without(imgs, removing) {
		ids = append(ids, img.ID)
	}
	layers, err := imageLayers(ids)
	if err != nil {
		return
	}
	kept := map[string]bool{}
	for _, img := range without(imgs, removing) {
		for _, l := range layers[img.ID] {
			kept[l] = true
		}
	}

	// An image tagged more than once frees on its last tag, which is where the
	// estimate already sits.
	last := map[string]int{}
	for i, t := range targets {
		if t.image != "" {
			last[t.image] = i
		}
	}
	order := make([]int, 0, len(last))
	for _, i := range last {
		order = append(order, i)
	}
	sort.Slice(order, func(a, b int) bool {
		la, lb := len(layers[targets[order[a]].image]), len(layers[targets[order[b]].image])
		if la != lb {
			return la < lb
		}
		return order[a] < order[b]
	})

	credited := map[string]bool{}
	for _, i := range order {
		id := targets[i].image
		history, err := historySizes(id)
		if err != nil {
			continue
		}
		sizes, ok := layerSizesFrom(layers[id], history)
		if !ok {
			continue
		}
		var freed int64
		for l, b := range sizes {
			if !kept[l] && !credited[l] {
				credited[l] = true
				freed += b
			}
		}
		targets[i].Bytes = freed
	}
}

// targetImages is the set of image IDs the targets would remove.
func targetImages(targets []Target) map[string]bool {
	ids := map[string]bool{}
	for _, t := range targets {
		if t.image != "" {
			ids[t.image] = true
		}
	}
	return ids
}

// without returns the images whose ID is not in ids.
func without(imgs []image, ids map[string]bool) []image {
	out := make([]image, 0, len(imgs))
	for _, img := range imgs {
		if !ids[img.ID] {
			out = append(out, img)
		}
	}
	return out
}
