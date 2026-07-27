import { useEffect, useMemo, useRef, useState } from "react";
import { createRoot } from "react-dom/client";
import { ChevronDown, X } from "lucide-react";
import akitaIdle from "./assets/pets/akita/akita_idle_8fps.gif";
import akitaRun from "./assets/pets/akita/akita_run_8fps.gif";
import akitaWalk from "./assets/pets/akita/akita_walk_8fps.gif";
import akitaSwipe from "./assets/pets/akita/akita_swipe_8fps.gif";
import akitaLie from "./assets/pets/akita/akita_lie_8fps.gif";
import akitaBall from "./assets/pets/akita/akita_with_ball_8fps.gif";
import "./pet.css";

type PetPhase = "idle" | "thinking" | "working" | "input" | "approval" | "done" | "error";

type PetSession = {
  tabId: string;
  title: string;
  status: string;
  phase: PetPhase;
  running?: boolean;
  waiting?: boolean;
};

type PetState = {
  title: string;
  status: string;
  phase?: PetPhase;
  tabId?: string;
  running?: boolean;
  waiting?: boolean;
  activeCount?: number;
  attentionCount?: number;
  updatedAt?: number;
  petId?: string;
  sessions?: PetSession[];
};

type PetPack = {
  id: string;
  displayName: string;
  description: string;
  kind: "gif-set" | "codex-spritesheet";
  builtIn?: boolean;
  assetUrl?: string;
};

type PetBindings = {
  ReadState(): Promise<PetState>;
  ReadCommand(): Promise<string>;
  ListDesktopPetPacks(): Promise<PetPack[]>;
  OpenMainWindow(tabId: string): Promise<void>;
  CloseDesktopPet(): Promise<void>;
};

function bindings(): PetBindings | undefined {
  return (window as typeof window & {
    go?: { main?: { DesktopPetApp?: PetBindings } };
  }).go?.main?.DesktopPetApp;
}

const akitaAnimations: Record<PetPhase, string> = {
  idle: akitaIdle,
  thinking: akitaWalk,
  working: akitaRun,
  input: akitaLie,
  approval: akitaSwipe,
  done: akitaBall,
  error: akitaLie,
};

const spriteRows: Record<PetPhase, number> = {
  idle: 0,
  thinking: 8,
  working: 7,
  input: 6,
  approval: 3,
  done: 4,
  error: 5,
};

function normalizePhase(state: PetState): PetPhase {
  if (state.phase) return state.phase;
  if (state.waiting) return "input";
  if (state.running) return "working";
  return "idle";
}

function DesktopPet() {
  const [state, setState] = useState<PetState>({ title: "Reasonix", status: "等待任务", phase: "idle", petId: "akita" });
  const [packs, setPacks] = useState<PetPack[]>([]);
  const [collapsed, setCollapsed] = useState(false);
  const [celebrating, setCelebrating] = useState(false);
  const previousRunning = useRef(false);

  useEffect(() => {
    let cancelled = false;
    void bindings()?.ListDesktopPetPacks().then((next) => {
      if (!cancelled) setPacks(next);
    });
    const refresh = () => {
      void bindings()?.ReadState().then((next) => {
        if (cancelled) return;
        const isRunning = Boolean(next.running || next.sessions?.some((session) => session.running));
        if (previousRunning.current && !isRunning && normalizePhase(next) === "idle") {
          setCelebrating(true);
          window.setTimeout(() => setCelebrating(false), 3600);
        }
        previousRunning.current = isRunning;
        setState(next);
      });
      void bindings()?.ReadCommand().then((command) => {
        if (cancelled || !command) return;
        if (command === "hide" || command === "close") void bindings()?.CloseDesktopPet();
        if (command === "collapse") setCollapsed(true);
        if (command === "expand" || command === "show") setCollapsed(false);
      });
    };
    refresh();
    const timer = window.setInterval(refresh, 500);
    return () => {
      cancelled = true;
      window.clearInterval(timer);
    };
  }, []);

  const phase = celebrating ? "done" : normalizePhase(state);
  const pack = useMemo(() => packs.find((item) => item.id === state.petId) ?? packs[0], [packs, state.petId]);
  const sessionStack = useMemo(() => {
    const priority = (session: PetSession) =>
      session.phase === "approval" ? 6 :
      session.phase === "input" ? 5 :
      session.phase === "error" ? 4 :
      session.running ? 3 :
      session.tabId === state.tabId ? 2 : 1;
    const relevant = (state.sessions ?? []).filter(
      (session) => session.running || session.waiting || session.tabId === state.tabId,
    );
    const ordered = [...relevant].sort((left, right) => priority(right) - priority(left));
    return {
      items: ordered.slice(0, 3),
      overflow: Math.max(0, ordered.length - 3),
    };
  }, [state.sessions, state.tabId]);
  const openMain = (tabId = state.tabId ?? "") => void bindings()?.OpenMainWindow(tabId);

  return (
    <main className={`pet-stage pet-stage--${phase}${collapsed ? " pet-stage--collapsed" : ""}`}>
      {!collapsed && sessionStack.items.length <= 1 && (
        <button className="pet-status-card pet-status-card--single" type="button" onClick={() => openMain()} title="点击返回当前对话">
          <span className={`pet-status-card__dot pet-status-card__dot--${phase}`} aria-hidden="true" />
          <span className="pet-status-card__copy">
            <strong>{state.title || "Reasonix"}</strong>
            <span className="pet-status-card__status">{state.status || "等待任务"}</span>
          </span>
        </button>
      )}

      {!collapsed && sessionStack.items.length > 1 && (
        <section className="pet-card-stack" aria-label="Reasonix 多任务状态">
          {sessionStack.items.map((session, index) => (
            <button
              key={session.tabId}
              className={`pet-status-card pet-status-card--stacked pet-status-card--${session.phase}`}
              style={{
                ["--stack-x" as string]: `${index * 8}px`,
                ["--stack-y" as string]: `${index * 11}px`,
                ["--stack-open-y" as string]: `${index * 73}px`,
                ["--stack-scale" as string]: 1 - index * 0.025,
                ["--stack-open-scale" as string]: 1 - index * 0.015,
                zIndex: sessionStack.items.length - index,
              }}
              type="button"
              onClick={() => openMain(session.tabId)}
              title={`打开：${session.title}`}
            >
              <span className={`pet-status-card__dot pet-status-card__dot--${session.phase}`} aria-hidden="true" />
              <span className="pet-status-card__copy">
                <strong>{session.title}</strong>
                <span className="pet-status-card__status">{session.status}</span>
              </span>
              {index === 0 && sessionStack.overflow > 0 && (
                <span className="pet-status-card__overflow">+{sessionStack.overflow}</span>
              )}
            </button>
          ))}
        </section>
      )}

      <button className="pet-character" type="button" onClick={() => openMain()} title="返回 Reasonix 对话">
        {pack?.kind === "codex-spritesheet" && pack.assetUrl ? (
          <span
            key={`${pack.id}-${phase}`}
            className="pet-sprite"
            style={{
              backgroundImage: `url(${pack.assetUrl})`,
              backgroundPositionY: `${spriteRows[phase] * -208}px`,
            }}
          />
        ) : (
          <img key={phase} src={akitaAnimations[phase]} alt="Reasonix 动画桌宠" draggable={false} />
        )}
        {(phase === "approval" || phase === "input" || phase === "error") && <span className="pet-alert" aria-hidden="true">!</span>}
      </button>

      <div className="pet-actions">
        <button type="button" onClick={() => setCollapsed((value) => !value)} title={collapsed ? "展开状态" : "收起状态"}>
          <ChevronDown className={collapsed ? "pet-action--collapsed" : ""} size={18} />
        </button>
        <button type="button" onClick={() => void bindings()?.CloseDesktopPet()} title="关闭桌宠">
          <X size={16} />
        </button>
      </div>
    </main>
  );
}

const root = document.getElementById("pet-root");
if (!root) throw new Error("missing #pet-root");
createRoot(root).render(<DesktopPet />);
