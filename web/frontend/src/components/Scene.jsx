import React from 'react';

/**
 * Atmospheric backdrop. All scene images render at fixed pixel sizes —
 * narrow viewports crop them at the edges (like the bg cover), they do not
 * shrink. Z-order back-to-front: bg → ghosts → table.
 */
export default function Scene() {
  return (
    <div className="scene" aria-hidden="true">
      <div className="scene-bg" />
      <img className="ghost ghost-pos-center ghost-fade-a" src="/ghost-behind.png" alt="" />
      <img className="ghost ghost-pos-left ghost-fade-b" src="/ghost-side.png" alt="" />
      <img className="ghost ghost-pos-right ghost-fade-c" src="/ghost-side.png" alt="" />
      <img className="scene-table" src="/table.png" alt="" />
    </div>
  );
}
