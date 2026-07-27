import { useEffect, useState } from "react";
import { createRoot } from "react-dom/client";
import { ChevronDown, X } from "lucide-react";
import cat from "./assets/fluent-cat.svg";
import "./pet.css";

type PetState = {
  title: string;
  status: string;
  running?: boolean;
  waiting?: boolean;
};

type PetBindings = {
  ReadState(): Promise<PetState>;
  OpenMainWindow(): Promise<void>;
  CloseDesktopPet(): Promise<void>;
};

function bindings(): PetBindings | undefined {
  return (window as typeof window & {
    go?: { main?: { DesktopPetApp?: PetBindings } };
  }).go?.main?.DesktopPetApp;
}

function DesktopPet() {
  const [state, setState] = useState<PetState>({ title: "Reasonix", status: "等待任务" });
  const [collapsed, setCollapsed] = useState(false);

  useEffect(() => {
    let cancelled = false;
    const refresh = () => {
      void bindings()?.ReadState().then((next) => {
        if (!cancelled) setState(next);
      });
    };
    refresh();
    const timer = window.setInterval(refresh, 600);
    return () => {
      cancelled = true;
      window.clearInterval(timer);
    };
  }, []);

  const openMain = () => void bindings()?.OpenMainWindow();

  return (
    <main className={`pet-stage${collapsed ? " pet-stage--collapsed" : ""}`}>
      {!collapsed && (
        <button className="pet-bubble" type="button" onClick={openMain} title="点击返回对话">
          <strong>{state.title || "Reasonix"}</strong>
          <span className={state.waiting ? "pet-bubble__status pet-bubble__status--waiting" : "pet-bubble__status"}>
            {state.status || (state.running ? "正在处理任务" : "等待任务")}
          </span>
        </button>
      )}

      <div className={`pet-character${state.running ? " pet-character--running" : ""}`}>
        <button className="pet-character__button" type="button" onClick={openMain} title="返回 Reasonix 对话">
          <img src={cat} alt="Reasonix 桌宠猫" draggable={false} />
        </button>
      </div>

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
