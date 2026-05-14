import React, { createContext, useContext, useEffect, useState, useCallback, useRef } from "react";

export type EventType = "new_attack" | "new_log" | "node_status" | "stats_update";

export interface WSEvent {
  type: EventType;
  timestamp: string;
  data: any;
}

export interface AttackEventData {
  id: number;
  node_id: string;
  service_id: number;
  client_ip: string;
  method: string;
  path: string;
  attack_type: string;
  attack_detail: string;
  user_agent: string;
  honeypot_type: number;
  created_at: string;
}

export interface StatsEventData {
  total_requests: number;
  attack_count: number;
  online_nodes: number;
  running_services: number;
}

interface WSContextValue {
  connected: boolean;
  lastEvent: WSEvent | null;
  subscribe: (types: EventType[]) => void;
  unsubscribe: () => void;
}

const WSContext = createContext<WSContextValue>({
  connected: false,
  lastEvent: null,
  subscribe: () => {},
  unsubscribe: () => {},
});

export function useWebSocket() {
  return useContext(WSContext);
}

export function WebSocketProvider({ children }: { children: React.ReactNode }) {
  const [connected, setConnected] = useState(false);
  const [lastEvent, setLastEvent] = useState<WSEvent | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const subscribedTypesRef = useRef<EventType[]>([]);

  const connect = useCallback(() => {
    const token = localStorage.getItem("token");
    if (!token) return;

    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const wsUrl = `${protocol}//${window.location.host}/ws/events?token=${encodeURIComponent(token)}`;

    try {
      const ws = new WebSocket(wsUrl);
      wsRef.current = ws;

      ws.onopen = () => {
        setConnected(true);
        if (reconnectTimerRef.current) {
          clearTimeout(reconnectTimerRef.current);
          reconnectTimerRef.current = null;
        }
        if (subscribedTypesRef.current.length > 0) {
          ws.send(JSON.stringify({
            type: "subscribe",
            payload: { types: subscribedTypesRef.current }
          }));
        }
      };

      ws.onmessage = (event) => {
        try {
          const msg = JSON.parse(event.data);
          const wsEvent: WSEvent = {
            type: msg.type,
            timestamp: msg.timestamp,
            data: msg.data
          };
          setLastEvent(wsEvent);
        } catch {}
      };

      ws.onerror = () => {};

      ws.onclose = () => {
        setConnected(false);
        wsRef.current = null;
        reconnectTimerRef.current = setTimeout(connect, 5000);
      };
    } catch {}
  }, []);

  useEffect(() => {
    connect();
    return () => {
      if (reconnectTimerRef.current) {
        clearTimeout(reconnectTimerRef.current);
      }
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, [connect]);

  const subscribe = useCallback((types: EventType[]) => {
    subscribedTypesRef.current = types;
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({
        type: "subscribe",
        payload: { types }
      }));
    }
  }, []);

  const unsubscribe = useCallback(() => {
    subscribedTypesRef.current = [];
  }, []);

  return (
    <WSContext.Provider value={{ connected, lastEvent, subscribe, unsubscribe }}>
      {children}
    </WSContext.Provider>
  );
}