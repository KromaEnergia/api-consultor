"use client";
import { useEffect, useState } from "react";
import ConsultantCard from "@/components/consultor/Perfil/consultorCard";
import ContratoNegociacoesCard from "@/components/consultor/negocios/ContratoNegociacoesCard";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Portal do Consultor",
  description: "Kroma Energia - Portal do Consultor",
};

// Definição dos tipos
interface Negociacao {
  negociacaoId: number;
  consultorId: number;
  valor: number;
  inicioSuprimento: string;
  fimSuprimento: string;
  valorIntegral: boolean;
}
interface Contrato {
  id: number;
  negociacaoId: number;
  consultorId: number;
  valor: number;
  inicioSuprimento: string;
  fimSuprimento: string;
  valorIntegral: boolean;
}
interface Consultor {
  id: number;
  nome: string;
  sobrenome: string;
  email: string;
  telefone: string;
  cnpj: string;
  foto: string;
  isAdmin: boolean;
  negociacoes: Negociacao[];
  contratos: Contrato[];
}

export default function ConsultantPage() {
  const [consultor, setConsultor] = useState<Consultor | null>(null);
  const [negociacoes, setNegociacoes] = useState<Negociacao[]>([]);
  const [contratos, setContratos] = useState<Contrato[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    (async () => {
      try {
        // 1) Autentica e obtém token (salvando no localStorage)
        let jwt = localStorage.getItem("jwtToken");
        if (!jwt) {
          const loginRes = await fetch("/api/login", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ login: "hitalloo", password: "123" }),
          });
          if (!loginRes.ok) throw new Error("Falha ao autenticar");
          const { token } = await loginRes.json();
          jwt = token;
          localStorage.setItem("jwtToken", jwt);
        }

        // 2) Busca dados do consultor via endpoint de resumo
        const res = await fetch(`/api/consultores/9/resumo`, {
          headers: { Authorization: `Bearer ${jwt}` },
        });
        if (!res.ok) throw new Error("Falha ao carregar resumo do consultor");
        // espera objeto com campos do consultor e arrays negociacoes e contratos
        const summary = await res.json() as Consultor;
        setConsultor(summary);
        setNegociacoes(summary.negociacoes);
        setContratos(summary.contratos);(data.contratos);
      } catch (err: any) {
        setError(err.message);
      }
    })();
  }, []);

  if (error) return <div className="text-red-500">Erro: {error}</div>;
  if (!consultor) return <div>Carregando...</div>;

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold">Portal do Consultor</h1>
      {/* Perfil */}
      <ConsultantCard
        nome={consultor.nome}
        sobrenome={consultor.sobrenome}
        cnpj={consultor.cnpj}
        foto={consultor.foto}
      />
      {/* Lista de negociações */}
      <div>
        <h2 className="text-xl font-medium">Negociações</h2>
        <ul className="space-y-4">
          {negociacoes.map((n) => (
            <li key={n.negociacaoId}>
              <ContratoNegociacoesCard
                valor={n.valor}
                valorIntegral={n.valorIntegral}
              />
            </li>
          ))}
        </ul>
      </div>
      {/* Lista de contratos */}
      <div>
        <h2 className="text-xl font-medium">Contratos</h2>
        <ul className="space-y-4">
          {contratos.map((c) => (
            <li key={c.id}>
              <ContratoNegociacoesCard
                valor={c.valor}
                valorIntegral={c.valorIntegral}
              />
            </li>
          ))}
        </ul>
      </div>
    </div>
  );
}
