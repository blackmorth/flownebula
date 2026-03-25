import { describe, expect, it } from 'vitest'
import buildTree from './buildTree'

describe('buildTree', () => {
  it('builds a recursive tree from payload nodes and edges', () => {
    const payload = {
      root: 'n1',
      nodes: {
        n1: { nodeId: 'root', inclusive_cost: { wt: 20 } },
        n2: { nodeId: 'childA', inclusive_cost: { wt: 15 } },
        n3: { nodeId: 'childB', inclusive_cost: { wt: 5 } },
      },
      edges: {
        e1: { caller: 'n1', callee: 'n2' },
        e2: { caller: 'n2', callee: 'n3' },
      },
    }

    expect(buildTree(payload)).toEqual({
      id: 'n1',
      name: 'root',
      cost: 20,
      children: [
        {
          id: 'n2',
          name: 'childA',
          cost: 15,
          children: [
            {
              id: 'n3',
              name: 'childB',
              cost: 5,
              children: [],
            },
          ],
        },
      ],
    })
  })

  it('stops building when a cycle is detected', () => {
    const payload = {
      root: 'n1',
      nodes: {
        n1: { nodeId: 'root', inclusive_cost: { wt: 20 } },
        n2: { nodeId: 'childA', inclusive_cost: { wt: 15 } },
      },
      edges: {
        e1: { caller: 'n1', callee: 'n2' },
        e2: { caller: 'n2', callee: 'n2' },
      },
    }

    expect(buildTree(payload)).toEqual({
      id: 'n1',
      name: 'root',
      cost: 20,
      children: [
        {
          id: 'n2',
          name: 'childA',
          cost: 15,
          children: [
            {
              id: 'n2',
              name: 'childA (recursion)',
              cost: 0,
              children: [],
            },
          ],
        },
      ],
    })
  })
})
