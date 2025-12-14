"use client";

import useSWR from "swr";

type Score = {
  id: number;
  artist: string;
  charter?: string | null;
  total_score?: number | null;
  stars_achieved?: number | null;
  created_at: string;
};

type Artist = {
  id: number;
  name: string;
  created_at: string;
};

type Song = {
  id: number;
  name: string;
  artist_id?: number | null;
  charters: string[];
  created_at: string;
};

const fetcher = async (url: string) => {
  const response = await fetch(url);
  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(`HTTP ${response.status}: ${errorText || response.statusText}`);
  }
  return response.json();
};

const apiBase = process.env.NEXT_PUBLIC_API_BASE || "http://localhost:8080";

export default function Page() {
  const { data: scores, error: scoresError, isLoading: scoresLoading } = useSWR<Score[]>(
    `${apiBase}/scores?limit=50`,
    fetcher
  );

  const { data: artists, error: artistsError, isLoading: artistsLoading } = useSWR<Artist[]>(
    `${apiBase}/artists?limit=100`,
    fetcher
  );

  const { data: songs, error: songsError, isLoading: songsLoading } = useSWR<Song[]>(
    `${apiBase}/songs?limit=100`,
    fetcher
  );

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: "2rem" }}>
      {/* Scores Section */}
      <div className="card">
        <h2>Recent Scores</h2>
        {scoresError && (
          <div style={{ color: "tomato" }}>
            <p>Failed to load scores</p>
            <p style={{ fontSize: "0.9em", marginTop: "0.5rem" }}>
              {scoresError instanceof Error ? scoresError.message : String(scoresError)}
            </p>
          </div>
        )}
        {scoresLoading && <p className="muted">Loading scores…</p>}
        {scores && (
          <table>
            <thead>
              <tr>
                <th>Artist</th>
                <th>Charter</th>
                <th>Total Score</th>
                <th>Stars</th>
                <th>Date</th>
              </tr>
            </thead>
            <tbody>
              {scores.length === 0 ? (
                <tr>
                  <td colSpan={5} style={{ textAlign: "center", color: "#8b949e" }}>
                    No scores yet
                  </td>
                </tr>
              ) : (
                scores.map((s) => (
                  <tr key={s.id}>
                    <td>{s.artist}</td>
                    <td>{s.charter ?? "—"}</td>
                    <td>{s.total_score?.toLocaleString() ?? "—"}</td>
                    <td>{s.stars_achieved ?? "—"}</td>
                    <td>{new Date(s.created_at).toLocaleString()}</td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        )}
      </div>

      {/* Artists Section */}
      <div className="card">
        <h2>Artists</h2>
        {artistsError && (
          <div style={{ color: "tomato" }}>
            <p>Failed to load artists</p>
            <p style={{ fontSize: "0.9em", marginTop: "0.5rem" }}>
              {artistsError instanceof Error ? artistsError.message : String(artistsError)}
            </p>
          </div>
        )}
        {artistsLoading && <p className="muted">Loading artists…</p>}
        {artists && (
          <table>
            <thead>
              <tr>
                <th>ID</th>
                <th>Name</th>
                <th>Created</th>
              </tr>
            </thead>
            <tbody>
              {artists.length === 0 ? (
                <tr>
                  <td colSpan={3} style={{ textAlign: "center", color: "#8b949e" }}>
                    No artists yet
                  </td>
                </tr>
              ) : (
                artists.map((a) => (
                  <tr key={a.id}>
                    <td>{a.id}</td>
                    <td>{a.name}</td>
                    <td>{new Date(a.created_at).toLocaleString()}</td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        )}
      </div>

      {/* Songs Section */}
      <div className="card">
        <h2>Songs</h2>
        {songsError && (
          <div style={{ color: "tomato" }}>
            <p>Failed to load songs</p>
            <p style={{ fontSize: "0.9em", marginTop: "0.5rem" }}>
              {songsError instanceof Error ? songsError.message : String(songsError)}
            </p>
          </div>
        )}
        {songsLoading && <p className="muted">Loading songs…</p>}
        {songs && (
          <table>
            <thead>
              <tr>
                <th>ID</th>
                <th>Name</th>
                <th>Artist ID</th>
                <th>Charters</th>
                <th>Created</th>
              </tr>
            </thead>
            <tbody>
              {songs.length === 0 ? (
                <tr>
                  <td colSpan={5} style={{ textAlign: "center", color: "#8b949e" }}>
                    No songs yet
                  </td>
                </tr>
              ) : (
                songs.map((s) => (
                  <tr key={s.id}>
                    <td>{s.id}</td>
                    <td>{s.name}</td>
                    <td>{s.artist_id ?? "—"}</td>
                    <td>
                      {s.charters && s.charters.length > 0
                        ? s.charters.join(", ")
                        : "—"}
                    </td>
                    <td>{new Date(s.created_at).toLocaleString()}</td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}
