import { Nav } from "@/_components/Nav";
import { Metadata } from "next";
import Markdown from "react-markdown";
import { pages } from "@/_constants/docs.pages";
import remarkGfm from "remark-gfm";

type Params = {
  slug: keyof typeof pages;
};

type Props = {
  params: Promise<Params>;
};

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const slug = (await params).slug;
  return {
    title: pages[slug].title,
  };
}

export default async function DocsPage({ params }: { params: Promise<Params> }) {
  const slug = (await params).slug;
  const definition = pages[slug];

  return (
    <>
      <Nav />
      <div className="prose max-w-4xl mx-auto px-8 my-8">
        <Markdown remarkPlugins={[remarkGfm]}>{definition.contents()}</Markdown>
      </div>
    </>
  );
}

export async function generateStaticParams(): Promise<Params[]> {
  return Object.keys(pages).map((slug) => ({
    slug: slug as keyof typeof pages,
  }));
}
